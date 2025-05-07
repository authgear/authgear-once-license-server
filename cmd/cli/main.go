package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/joho/godotenv"
	slogmulti "github.com/samber/slog-multi"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-once-license-server/pkg/emailtemplate"
	"github.com/authgear/authgear-once-license-server/pkg/httpmiddleware"
	"github.com/authgear/authgear-once-license-server/pkg/installationscript"
	"github.com/authgear/authgear-once-license-server/pkg/keygen"
	"github.com/authgear/authgear-once-license-server/pkg/slogging"
	"github.com/authgear/authgear-once-license-server/pkg/smtp"
	pkgstripe "github.com/authgear/authgear-once-license-server/pkg/stripe"
)

const indexHTML = `<!DOCTYPE html>
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
	<form action="/v1/stripe/checkout" method="POST">
		<button type="submit">Checkout</button>
	</form>
</body>
</html>
`

var jsonResponseBadRequest = map[string]any{
	"error": map[string]any{
		"code": "bad_request",
	},
}

var jsonResponseInternalServerError = map[string]any{
	"error": map[string]any{
		"code": "internal_server_error",
	},
}

var jsonResponseLicenseKeyNotFound = map[string]any{
	"error": map[string]any{
		"code": "license_key_not_found",
	},
}

var jsonResponseLicenseKeyAlreadyActivated = map[string]any{
	"error": map[string]any{
		"code": "license_key_already_activated",
	},
}

func NewLicenseResponse(l *keygen.LicenseID) map[string]any {
	return map[string]any{
		"data": l,
	}
}

var rootCmd = &cobra.Command{
	Use: "authgear-once-license-server",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server at port 8200",
	RunE: func(cmd *cobra.Command, args []string) error {
		mux := http.NewServeMux()
		cors := httpmiddleware.CORSMiddleware(os.Getenv("AUTHGEAR_ONCE_CORS_ALLOWED_ORIGINS"))
		maxbytes := httpmiddleware.MaxBytesMiddleware(100 * 1000) // 100KB

		mux.HandleFunc("GET /{$}", Handler_root)
		mux.HandleFunc("GET /install/{license_key}", Handler_install)
		mux.HandleFunc("/v1/license/activate", MakeHandler_v1_license(keygen.ActivateLicense))
		mux.HandleFunc("/v1/license/check", MakeHandler_v1_license(keygen.CheckLicense))
		mux.HandleFunc("/v1/stripe/checkout", Handler_v1_stripe_checkout)
		mux.HandleFunc("/v1/stripe/webhook", Handler_v1_stripe_webhook)

		ctx := cmd.Context()
		logger := slogging.GetLogger(ctx)
		server := &http.Server{
			Addr:    ":8200",
			Handler: maxbytes(cors(mux)),
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		}

		err := server.ListenAndServe()
		if err != nil {
			slogging.Error(ctx, logger, "failed to start server",
				"error", err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

type dependenciesKeyType struct{}

var dependenciesKey = dependenciesKeyType{}

type Dependencies struct {
	HTTPClient                                          *http.Client
	StripeClient                                        *client.API
	SMTPDialer                                          *gomail.Dialer
	SMTPSender                                          string
	StripeCheckoutSessionSuccessURL                     string
	StripeCheckoutSessionCancelURL                      string
	StripeCheckoutSessionPriceID                        string
	StripeWebhookSigningSecret                          string
	AUTHGEAR_ONCE_PUBLIC_URL_SCHEME                     string
	AUTHGEAR_ONCE_ONCE_COMMAND_DOWNLOAD_URL_GO_TEMPLATE string
	AUTHGEAR_ONCE_ONCE_COMMAND_IMAGE_OVERRIDE           string
	KeygenConfig                                        keygen.KeygenConfig
}

func ConstructFullURL(r *http.Request) *url.URL {
	ctx := r.Context()
	deps := GetDependencies(ctx)
	scheme := "https"
	if deps.AUTHGEAR_ONCE_PUBLIC_URL_SCHEME == "http" {
		scheme = "http"
	}
	host := r.Host

	u := *r.URL
	u.Scheme = scheme
	u.Host = host
	return &u
}

func WriteJSON(w http.ResponseWriter, jsonBody any, statusCode int) {
	jsonBytes, err := json.Marshal(jsonBody)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonBytes)))
	w.WriteHeader(statusCode)
	w.Write(jsonBytes)
}

func GetDependencies(ctx context.Context) Dependencies {
	return ctx.Value(dependenciesKey).(Dependencies)
}

func Handler_root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}

func Handler_install(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	deps := GetDependencies(ctx)
	logger := slogging.GetLogger(ctx)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	licenseKey := r.PathValue("license_key")
	uname_s := r.FormValue("uname_s")
	uname_m := r.FormValue("uname_m")

	switch {
	case uname_s == "" || uname_m == "":
		// uname_s or uname_m is unspecified.
		// This is the case of the link in the email.
		// In this case, we return a shell script that is supposed to be run by a oneliner.

		u := ConstructFullURL(r)

		script, err := installationscript.Render(installationscript.RenderOptions{
			DownloadURL:   u.String(),
			LicenseKey:    licenseKey,
			ImageOverride: deps.AUTHGEAR_ONCE_ONCE_COMMAND_IMAGE_OVERRIDE,
		})
		if err != nil {
			slogging.Error(ctx, logger, "failed to render installation shell script",
				"error", err)
			http.Error(w, "failed to render installation shell script", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Cache-Control", "no-store")
		w.Write([]byte(script))
	default:
		downloadURL, err := installationscript.RenderDownloadURL(deps.AUTHGEAR_ONCE_ONCE_COMMAND_DOWNLOAD_URL_GO_TEMPLATE, installationscript.RenderDownloadURLOptions{
			Uname_s: uname_s,
			Uname_m: uname_m,
		})
		if err != nil {
			slogging.Error(ctx, logger, "failed to render download url",
				"error", err)
			http.Error(w, "failed to render download url", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, downloadURL, http.StatusSeeOther)
	}
}

func MakeHandler_v1_license(f func(ctx context.Context, httpClient *http.Client, opts keygen.LicenseOptions) (*keygen.LicenseID, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		ctx := r.Context()
		logger := slogging.GetLogger(ctx)
		deps := GetDependencies(ctx)

		err := r.ParseForm()
		if err != nil {
			WriteJSON(w, jsonResponseBadRequest, http.StatusBadRequest)
			return
		}

		licenseKey := r.FormValue("license_key")
		fingerprint := r.FormValue("fingerprint")
		if licenseKey == "" || fingerprint == "" {
			WriteJSON(w, jsonResponseBadRequest, http.StatusBadRequest)
			return
		}

		licenseID, err := f(ctx, deps.HTTPClient, keygen.LicenseOptions{
			KeygenConfig: deps.KeygenConfig,
			LicenseKey:   licenseKey,
			Fingerprint:  fingerprint,
		})
		if err != nil {
			switch {
			case errors.Is(err, keygen.ErrLicenseKeyNotFound):
				WriteJSON(w, jsonResponseLicenseKeyNotFound, http.StatusNotFound)
				return
			case errors.Is(err, keygen.ErrLicenseKeyAlreadyActivated):
				WriteJSON(w, jsonResponseLicenseKeyAlreadyActivated, http.StatusForbidden)
				return
			default:
				slogging.Error(ctx, logger, "unexpected error",
					"error", err)
				WriteJSON(w, jsonResponseInternalServerError, http.StatusInternalServerError)
				return
			}
		}
		if licenseID.StripeCustomerID != "" {
			customer, err := pkgstripe.GetCustomer(ctx, deps.StripeClient, licenseID.StripeCustomerID)
			if err != nil {
				slogging.Error(ctx, logger, "unexpected error",
					"error", err)
				WriteJSON(w, jsonResponseInternalServerError, http.StatusInternalServerError)
				return
			}

			licenseID.LicenseeEmail = &customer.Email
		}

		WriteJSON(w, NewLicenseResponse(licenseID), http.StatusOK)
	}
}

func Handler_v1_stripe_checkout(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx := r.Context()
	logger := slogging.GetLogger(ctx)
	deps := GetDependencies(ctx)
	stripeClient := deps.StripeClient

	checkoutSession, err := pkgstripe.NewCheckoutSession(ctx, stripeClient, &pkgstripe.CheckoutSessionParams{
		SuccessURL: deps.StripeCheckoutSessionSuccessURL,
		CancelURL:  deps.StripeCheckoutSessionCancelURL,
		PriceID:    deps.StripeCheckoutSessionPriceID,
	})
	if err != nil {
		slogging.Error(ctx, logger, "failed to create checkout session",
			"error", err)
		http.Error(w, "failed to create checkout session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, checkoutSession.URL, http.StatusSeeOther)
}

func Handler_v1_stripe_webhook(w http.ResponseWriter, r *http.Request) {
	// The API version must be 2025-03-31.basil
	defer r.Body.Close()

	ctx := r.Context()
	logger := slogging.GetLogger(ctx)
	deps := GetDependencies(ctx)

	e, err := pkgstripe.ConstructEvent(r, deps.StripeWebhookSigningSecret)
	if err != nil {
		slogging.Error(ctx, logger, "failed to construct webhook event",
			"error", err)
		if !pkgstripe.IsWebhookClientError(err) {
			http.Error(w, "failed to construct webhook event", http.StatusInternalServerError)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	switch e.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		b, err := json.Marshal(e)
		if err != nil {
			panic(err)
		}
		logger = logger.With("stripe_event_json_base64url", base64.RawURLEncoding.EncodeToString(b))

		checkoutSessionID, ok := pkgstripe.GetCheckoutSessionID(e)
		if !ok {
			slogging.Error(ctx, logger, "checkout session ID not found")
			http.Error(w, "checkout session ID not found", http.StatusInternalServerError)
			return
		}

		customerID, ok := pkgstripe.GetCustomerID(e)
		if !ok {
			slogging.Error(ctx, logger, "customer id not found")
			http.Error(w, "customer id not found", http.StatusInternalServerError)
			return
		}

		email, ok := pkgstripe.GetCustomerEmail(e)
		if !ok {
			slogging.Error(ctx, logger, "customer email not found")
			http.Error(w, "customer email not found", http.StatusInternalServerError)
			return
		}

		licenseKey, err := keygen.CreateLicenseKey(ctx, deps.HTTPClient, keygen.CreateLicenseKeyOptions{
			KeygenConfig:            deps.KeygenConfig,
			StripeCheckoutSessionID: checkoutSessionID,
			StripeCustomerID:        customerID,
		})
		if err != nil {
			slogging.Error(ctx, logger, "failed to create license key",
				"error", err)
			http.Error(w, "failed to create license key", http.StatusInternalServerError)
			return
		}

		u := ConstructFullURL(r)
		u.Path = fmt.Sprintf("/install/%v", licenseKey)

		htmlBody := emailtemplate.RenderInstallationEmail(emailtemplate.InstallationEmailData{
			InstallationOneliner: fmt.Sprintf(`/bin/sh -c "$(curl -fsSL %v)"`, u.String()),
		})

		opts := smtp.EmailOptions{
			Sender:   deps.SMTPSender,
			Subject:  "Installing Authgear once",
			HTMLBody: htmlBody,
			To:       email,
		}

		err = smtp.SendEmail(deps.SMTPDialer, opts)
		if err != nil {
			slogging.Error(ctx, logger, "failed to send email",
				"error", err)
			http.Error(w, "failed to send email", http.StatusInternalServerError)
		} else {
			slogging.Info(ctx, logger, "sent installation to checkout session",
				"checkout_session_id", checkoutSessionID,
				"customer_id", customerID)
			// Return 200 implicitly.
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("AUTHGEAR_ONCE_SENTRY_SDN"),
		AttachStacktrace: true,
		EnableTracing:    false,
	})
	if err != nil {
		panic(err)
	}
	defer sentry.Flush(2 * time.Second)

	stripeClient := pkgstripe.NewClient(os.Getenv("AUTHGEAR_ONCE_STRIPE_SECRET_KEY"))
	smtpPort, err := strconv.Atoi(os.Getenv("AUTHGEAR_ONCE_SMTP_PORT"))
	if err != nil {
		panic(err)
	}

	smtpDialer := smtp.NewDialer(smtp.NewDialerOptions{
		SMTPHost:     os.Getenv("AUTHGEAR_ONCE_SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("AUTHGEAR_ONCE_SMTP_USERNAME"),
		SMTPPassword: os.Getenv("AUTHGEAR_ONCE_SMTP_PASSWORD"),
	})

	dependencies := Dependencies{
		HTTPClient:                      &http.Client{},
		StripeClient:                    stripeClient,
		SMTPDialer:                      smtpDialer,
		SMTPSender:                      os.Getenv("AUTHGEAR_ONCE_SMTP_SENDER"),
		StripeCheckoutSessionSuccessURL: os.Getenv("AUTHGEAR_ONCE_STRIPE_CHECKOUT_SESSION_SUCCESS_URL"),
		StripeCheckoutSessionCancelURL:  os.Getenv("AUTHGEAR_ONCE_STRIPE_CHECKOUT_SESSION_CANCEL_URL"),
		StripeCheckoutSessionPriceID:    os.Getenv("AUTHGEAR_ONCE_STRIPE_CHECKOUT_SESSION_PRICE_ID"),
		StripeWebhookSigningSecret:      os.Getenv("AUTHGEAR_ONCE_STRIPE_WEBHOOK_SIGNING_SECRET"),
		AUTHGEAR_ONCE_PUBLIC_URL_SCHEME: os.Getenv("AUTHGEAR_ONCE_PUBLIC_URL_SCHEME"),
		AUTHGEAR_ONCE_ONCE_COMMAND_DOWNLOAD_URL_GO_TEMPLATE: os.Getenv("AUTHGEAR_ONCE_ONCE_COMMAND_DOWNLOAD_URL_GO_TEMPLATE"),
		AUTHGEAR_ONCE_ONCE_COMMAND_IMAGE_OVERRIDE:           os.Getenv("AUTHGEAR_ONCE_ONCE_COMMAND_IMAGE_OVERRIDE"),
		KeygenConfig: keygen.KeygenConfig{
			Endpoint:   os.Getenv("AUTHGEAR_ONCE_KEYGEN_ENDPOINT"),
			AdminToken: os.Getenv("AUTHGEAR_ONCE_KEYGEN_ADMIN_TOKEN"),
			PolicyID:   os.Getenv("AUTHGEAR_ONCE_KEYGEN_POLICY_ID"),
		},
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, dependenciesKey, dependencies)

	textHandler := slog.NewTextHandler(os.Stderr, nil)
	sentryHandler := sentryslog.Option{
		Level:     slog.LevelError,
		AddSource: true,
	}.NewSentryHandler()
	handler := slogmulti.Fanout(textHandler, sentryHandler)
	logger := slog.New(handler)

	ctx = slogging.WithLogger(ctx, logger)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slogging.Error(ctx, logger, "root command completed with error",
			"error", err)
		os.Exit(1)
	}
}
