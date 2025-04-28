package keygen

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var ErrUnexpectedResponse = errors.New("unexpected response")

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

func patchRequest(r *http.Request) {
	// Keygen requires TLS.
	// We tell it is.
	r.Header.Set("X-Forwarded-Proto", "https")
}
