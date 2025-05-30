package keygen

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var ErrUnexpectedResponse = errors.New("unexpected response")
var ErrLicenseKeyNotFound = errors.New("license key not found")
var ErrLicenseKeyAlreadyActivated = errors.New("license key already activated")

type KeygenResponseError struct {
	DumpedResponse []byte
}

func (e *KeygenResponseError) Error() string {
	return fmt.Sprintf("keygen response: %v", base64.RawURLEncoding.EncodeToString(e.DumpedResponse))
}

type KeygenConfig struct {
	Endpoint   string
	AdminToken string
	PolicyID   string
}

type CreateLicenseKeyOptions struct {
	KeygenConfig            KeygenConfig
	StripeCheckoutSessionID string
	StripeCustomerID        string
}

func CreateLicenseKey(ctx context.Context, client *http.Client, opts CreateLicenseKeyOptions) (licenseKey string, err error) {
	u, err := url.JoinPath(opts.KeygenConfig.Endpoint, "/v1/licenses")
	if err != nil {
		return
	}

	reqBody := map[string]any{
		"data": map[string]any{
			"type": "license",
			"attributes": map[string]any{
				"metadata": map[string]any{
					"stripe_checkout_session_id": opts.StripeCheckoutSessionID,
					"stripe_customer_id":         opts.StripeCustomerID,
				},
			},
			"relationships": map[string]any{
				"policy": map[string]any{
					"data": map[string]any{
						"type": "policy",
						"id":   opts.KeygenConfig.PolicyID,
					},
				},
			},
		},
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return
	}
	patchRequest(req)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", opts.KeygenConfig.AdminToken))

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, &KeygenResponseError{DumpedResponse: dumpedResponse})
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var respBody map[string]any
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		if err != nil {
			return
		}

		licenseKey = respBody["data"].(map[string]any)["attributes"].(map[string]any)["key"].(string)
		return
	}

	err = ErrUnexpectedResponse
	return
}

type validateLicenseKeyOptions struct {
	KeygenConfig KeygenConfig
	LicenseKey   string
	Fingerprint  string
}

// LicenseID is an API object.
type LicenseID struct {
	ID          string     `json:"-"`
	ExpireAt    *time.Time `json:"expire_at"`
	IsActivated bool       `json:"is_activated"`
	IsExpired   bool       `json:"is_expired"`

	StripeCheckoutSessionID string  `json:"-"`
	StripeCustomerID        string  `json:"-"`
	LicenseeEmail           *string `json:"licensee_email"`
}

// validateLicenseKey returns the following errors:
// - ErrUnexpectedResponse
// - ErrLicenseKeyNotFound
// - ErrLicenseKeyAlreadyActivated
func validateLicenseKey(ctx context.Context, client *http.Client, opts validateLicenseKeyOptions) (licenseID *LicenseID, err error) {
	u, err := url.JoinPath(opts.KeygenConfig.Endpoint, "/v1/licenses/actions/validate-key")
	if err != nil {
		return
	}

	reqBody := map[string]any{
		"meta": map[string]any{
			"key": opts.LicenseKey,
			"scope": map[string]any{
				"fingerprint": opts.Fingerprint,
			},
		},
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return
	}
	patchRequest(req)
	// This request is supposed to be called by anyone with the license key, so admin token is not needed.

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, &KeygenResponseError{DumpedResponse: dumpedResponse})
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return parseValidateLicenseKeyResponseBody(resp.Body)
	}

	err = ErrUnexpectedResponse
	return
}

func parseValidateLicenseKeyResponseBody(r io.Reader) (licenseID *LicenseID, err error) {
	var respBody map[string]any
	err = json.NewDecoder(r).Decode(&respBody)
	if err != nil {
		return
	}

	meta := respBody["meta"].(map[string]any)
	meta_code := meta["code"].(string)

	if meta_code == "NOT_FOUND" {
		err = ErrLicenseKeyNotFound
		return
	}

	if meta_code == "FINGERPRINT_SCOPE_MISMATCH" {
		err = ErrLicenseKeyAlreadyActivated
		return
	}

	data, ok := respBody["data"].(map[string]any)
	if !ok || data == nil {
		err = ErrUnexpectedResponse
		return
	}
	attributes, ok := data["attributes"].(map[string]any)
	if !ok {
		err = ErrUnexpectedResponse
		return
	}

	licenseID = &LicenseID{
		ID: data["id"].(string),
	}

	if expiry, ok := attributes["expiry"].(string); ok {
		var expireAt time.Time
		expireAt, err = time.Parse(time.RFC3339, expiry)
		if err != nil {
			err = errors.Join(err, ErrUnexpectedResponse)
			return
		}
		licenseID.ExpireAt = &expireAt
	}
	if metadata, ok := attributes["metadata"].(map[string]any); ok {
		if stripeCheckoutSessionID, ok := metadata["stripeCheckoutSessionId"].(string); ok {
			licenseID.StripeCheckoutSessionID = stripeCheckoutSessionID
		}
		if stripeCustomerID, ok := metadata["stripeCustomerId"].(string); ok {
			licenseID.StripeCustomerID = stripeCustomerID
		}
	}

	switch meta_code {
	case "VALID":
		licenseID.IsActivated = true
		licenseID.IsExpired = false
	case "EXPIRED":
		licenseID.IsActivated = true
		licenseID.IsExpired = true
	case "NO_MACHINE":
		licenseID.IsActivated = false
		licenseID.IsExpired = false
	default:
		err = ErrUnexpectedResponse
		return
	}

	return
}

type createMachineOptions struct {
	KeygenConfig KeygenConfig
	LicenseKey   string
	LicenseID    string
	Fingerprint  string
}

// createMachine returns the following errors:
// - ErrUnexpectedResponse
// - ErrLicenseKeyAlreadyActivated
func createMachine(ctx context.Context, client *http.Client, opts createMachineOptions) (err error) {
	u, err := url.JoinPath(opts.KeygenConfig.Endpoint, "/v1/machines")
	if err != nil {
		return
	}

	reqBody := map[string]any{
		"data": map[string]any{
			"type": "machines",
			"attributes": map[string]any{
				"fingerprint": opts.Fingerprint,
			},
			"relationships": map[string]any{
				"license": map[string]any{
					"data": map[string]any{
						"type": "license",
						"id":   opts.LicenseID,
					},
				},
			},
		},
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return
	}
	patchRequest(req)
	// This endpoint requires license key Authorization
	req.Header.Set("Authorization", fmt.Sprintf("License %v", opts.LicenseKey))

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, &KeygenResponseError{DumpedResponse: dumpedResponse})
		}
	}()

	// This endpoint has two responses.
	// 1. "data" is present
	// 2. "errors" is present
	//
	// When "data" is present, the machine is created successfully.
	// The response looks like
	// {
	//     "data": {
	//         "id": "74c5d5d9-5e72-4883-a154-9a7689e19604",
	//         "type": "machines",
	//         "attributes": {
	//             "fingerprint": "fg1",
	//             "cores": null,
	//             "ip": null,
	//             "hostname": null,
	//             "platform": null,
	//             "name": null,
	//             "requireHeartbeat": false,
	//             "heartbeatStatus": "NOT_STARTED",
	//             "heartbeatDuration": 600,
	//             "maxProcesses": null,
	//             "lastCheckOut": null,
	//             "lastHeartbeat": null,
	//             "nextHeartbeat": null,
	//             "metadata": {},
	//             "created": "2025-04-28T07:17:26.725Z",
	//             "updated": "2025-04-28T07:17:26.725Z"
	//         },
	//         "relationships": {
	//             "account": {
	//                 "links": {
	//                     "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151"
	//                 },
	//                 "data": {
	//                     "type": "accounts",
	//                     "id": "87c9078c-6ce3-4c29-9fa2-e092a594b151"
	//                 }
	//             },
	//             "product": {
	//                 "links": {
	//                     "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/machines/74c5d5d9-5e72-4883-a154-9a7689e19604/product"
	//                 },
	//                 "data": {
	//                     "type": "products",
	//                     "id": "c0fbaab6-034e-48f3-8a78-70bb8c59e2cf"
	//                 }
	//             },
	//             "group": {
	//                 "links": {
	//                     "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/machines/74c5d5d9-5e72-4883-a154-9a7689e19604/group"
	//                 },
	//                 "data": null
	//             },
	//             "license": {
	//                 "links": {
	//                     "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/machines/74c5d5d9-5e72-4883-a154-9a7689e19604/license"
	//                 },
	//                 "data": {
	//                     "type": "licenses",
	//                     "id": "9d1e8df9-229f-4b5d-a207-945dcfa1e996"
	//                 }
	//             },
	//             "owner": {
	//                 "links": {
	//                     "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/machines/74c5d5d9-5e72-4883-a154-9a7689e19604/owner"
	//                 },
	//                 "data": null
	//             },
	//             "components": {
	//                 "links": {
	//                     "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/machines/74c5d5d9-5e72-4883-a154-9a7689e19604/components"
	//                 }
	//             },
	//             "processes": {
	//                 "links": {
	//                     "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/machines/74c5d5d9-5e72-4883-a154-9a7689e19604/processes"
	//                 }
	//             }
	//         },
	//         "links": {
	//             "self": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/machines/74c5d5d9-5e72-4883-a154-9a7689e19604"
	//         }
	//     }
	// }
	//
	// When "errors" is present,
	// The response looks like,
	// {
	//     "errors": [
	//         {
	//             "title": "Unprocessable resource",
	//             "detail": "has already been taken",
	//             "code": "FINGERPRINT_TAKEN",
	//             "source": {
	//                 "pointer": "/data/attributes/fingerprint"
	//             },
	//             "links": {
	//                 "about": "https://keygen.sh/docs/api/machines/#machines-object-attrs-fingerprint"
	//             }
	//         },
	//         {
	//             "title": "Unprocessable resource",
	//             "detail": "machine count has exceeded maximum allowed for license (1)",
	//             "code": "MACHINE_LIMIT_EXCEEDED",
	//             "source": {
	//                 "pointer": "/data"
	//             },
	//             "links": {
	//                 "about": "https://keygen.sh/docs/api/machines/#machines-object"
	//             }
	//         }
	//     ],
	//     "meta": {
	//         "id": "0c154eb4-53c1-4dd7-ab36-5d65a0baaa93"
	//     }
	// }

	var respBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return
	}

	errors, ok := respBody["errors"].([]any)
	if ok {
		for _, anyError := range errors {
			errorObj := anyError.(map[string]any)
			code := errorObj["code"].(string)
			if code == "MACHINE_LIMIT_EXCEEDED" {
				err = ErrLicenseKeyAlreadyActivated
				return
			}
		}
		// Otherwise it is an unknown response.
		err = ErrUnexpectedResponse
		return
	}
	// Otherwise we treat it as success.
	return
}

type LicenseOptions struct {
	KeygenConfig KeygenConfig
	LicenseKey   string
	Fingerprint  string
}

// ActivateLicense returns the following errors:
// - ErrUnexpectedResponse
// - ErrLicenseKeyNotFound
// - ErrLicenseKeyAlreadyActivated
func ActivateLicense(ctx context.Context, client *http.Client, opts LicenseOptions) (licenseID *LicenseID, err error) {
	// We first try to validate the license key.
	// If the license is activated, we can return early.
	licenseID, err = validateLicenseKey(ctx, client, validateLicenseKeyOptions{
		KeygenConfig: opts.KeygenConfig,
		LicenseKey:   opts.LicenseKey,
		Fingerprint:  opts.Fingerprint,
	})
	if err != nil {
		return
	}
	// Activate the same fingerprint is idempotent.
	if licenseID.IsActivated {
		return
	}
	// Otherwise, the license key is not activated yet.
	// Activate it by creating a machine.

	err = createMachine(ctx, client, createMachineOptions{
		KeygenConfig: opts.KeygenConfig,
		LicenseKey:   opts.LicenseKey,
		LicenseID:    licenseID.ID,
		Fingerprint:  opts.Fingerprint,
	})
	if err != nil {
		return
	}

	licenseID, err = validateLicenseKey(ctx, client, validateLicenseKeyOptions{
		KeygenConfig: opts.KeygenConfig,
		LicenseKey:   opts.LicenseKey,
		Fingerprint:  opts.Fingerprint,
	})
	if err != nil {
		return
	}
	if !licenseID.IsActivated {
		err = ErrUnexpectedResponse
		return
	}

	return
}

// CheckLicense returns the following errors:
// - ErrUnexpectedResponse
// - ErrLicenseKeyNotFound
// - ErrLicenseKeyAlreadyActivated
func CheckLicense(ctx context.Context, client *http.Client, opts LicenseOptions) (licenseID *LicenseID, err error) {
	licenseID, err = validateLicenseKey(ctx, client, validateLicenseKeyOptions{
		KeygenConfig: opts.KeygenConfig,
		LicenseKey:   opts.LicenseKey,
		Fingerprint:  opts.Fingerprint,
	})
	return
}

func patchRequest(r *http.Request) {
	// Keygen requires TLS.
	// We tell it is.
	r.Header.Set("X-Forwarded-Proto", "https")
}
