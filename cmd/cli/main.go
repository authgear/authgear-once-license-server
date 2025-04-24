package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	texttemplate "text/template"
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
	"github.com/authgear/authgear-once-license-server/pkg/slogging"
	"github.com/authgear/authgear-once-license-server/pkg/smtp"
	pkgstripe "github.com/authgear/authgear-once-license-server/pkg/stripe"
	"github.com/authgear/authgear-once-license-server/pkg/uname"
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

var installationShellScript *texttemplate.Template

func init() {
	t, err := texttemplate.New("").Parse(`#!/bin/sh
set -e

echo "Installing the Authgear ONCE command......"
echo "This script uses sudo, you will be prompted for authentication."
sudo true

download_url="{{ $.DownloadURL }}?uname_s=$(uname -s)&uname_m=$(uname -m)"
tmp_path="$(mktemp)"
curl -fsSL "$download_url" > "$tmp_path"
sudo mv "$tmp_path" /usr/local/bin/authgear-once
sudo chmod u+x /usr/local/bin/authgear-once

if [ "$(uname -s)" = "Darwin" ]; then
	/usr/local/bin/authgear-once setup "{{ $.LicenseKey }}"
else
	sudo /usr/local/bin/authgear-once setup "{{ $.LicenseKey }}"
fi
`)
	if err != nil {
		panic(err)
	}
	installationShellScript = t
}

var rootCmd = &cobra.Command{
	Use: "authgear-once-license-server",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server at port 8200",
	RunE: func(cmd *cobra.Command, args []string) error {
		mux := http.NewServeMux()
		cors := httpmiddleware.CORSMiddleware(os.Getenv("CORS_ALLOWED_ORIGINS"))
		maxbytes := httpmiddleware.MaxBytesMiddleware(100 * 1000) // 100KB

		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(indexHTML))
		})

		mux.HandleFunc("GET /install/{license_key}", func(w http.ResponseWriter, r *http.Request) {
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

				data := map[string]any{
					"DownloadURL": u.String(),
					"LicenseKey":  licenseKey,
				}
				var buf bytes.Buffer
				err = installationShellScript.Execute(&buf, data)
				if err != nil {
					slogging.Error(ctx, logger, "failed to render installation shell script",
						"error", err)
					http.Error(w, "failed to render installation shell script", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("Cache-Control", "no-store")
				w.Write(buf.Bytes())
			default:
				// Otherwise, we return 303 to download the binary.
				uname_s = uname.NormalizeUnameS(uname_s)
				uname_m = uname.NormalizeUnameM(uname_m)
				data := map[string]any{
					"Uname_s": uname_s,
					"Uname_m": uname_m,
				}
				var buf strings.Builder
				err = deps.AuthgearOnceDownloadURLGoTemplate.Execute(&buf, data)
				if err != nil {
					slogging.Error(ctx, logger, "failed to render download url",
						"error", err)
					http.Error(w, "failed to render download url", http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, buf.String(), http.StatusSeeOther)
			}
		})

		mux.HandleFunc("/v1/stripe/checkout", func(w http.ResponseWriter, r *http.Request) {
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
		})

		// The API version must be 2025-03-31.basil
		mux.HandleFunc("/v1/stripe/webhook", func(w http.ResponseWriter, r *http.Request) {
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
				logger = logger.With("stripe_event_json", string(b))

				id, ok := pkgstripe.GetCheckoutSessionID(e)
				if !ok {
					slogging.Error(ctx, logger, "checkout session ID not found")
					return
				}

				email, ok := pkgstripe.GetCustomerEmail(e)
				if !ok {
					slogging.Error(ctx, logger, "customer email not found")
					return
				}

				u := ConstructFullURL(r)
				// FIXME: generate license key.
				u.Path = fmt.Sprintf("/install/%v", "TODO")

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
				} else {
					slogging.Info(ctx, logger, "sent installation to checkout session",
						"checkout_session_id", id)
				}
			}
		})

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
	StripeClient                      *client.API
	SMTPDialer                        *gomail.Dialer
	SMTPSender                        string
	StripeCheckoutSessionSuccessURL   string
	StripeCheckoutSessionCancelURL    string
	StripeCheckoutSessionPriceID      string
	StripeWebhookSigningSecret        string
	PublicURLScheme                   string
	AuthgearOnceDownloadURLGoTemplate *texttemplate.Template
}

func ConstructFullURL(r *http.Request) *url.URL {
	ctx := r.Context()
	deps := GetDependencies(ctx)
	scheme := "https"
	if deps.PublicURLScheme == "http" {
		scheme = "http"
	}
	host := r.Host

	u := *r.URL
	u.Scheme = scheme
	u.Host = host
	return &u
}

func GetDependencies(ctx context.Context) Dependencies {
	return ctx.Value(dependenciesKey).(Dependencies)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		EnableTracing:    false,
	})
	if err != nil {
		panic(err)
	}
	defer sentry.Flush(2 * time.Second)

	stripeClient := pkgstripe.NewClient(os.Getenv("STRIPE_SECRET_KEY"))
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		panic(err)
	}

	smtpDialer := smtp.NewDialer(smtp.NewDialerOptions{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
	})

	authgearOnceDownloadURLGoTemplate, err := texttemplate.New("").Parse(os.Getenv("AUTHGEAR_ONCE_BINARY_DOWNLOAD_URL_GO_TEMPLATE"))
	if err != nil {
		panic(err)
	}

	dependencies := Dependencies{
		StripeClient:                      stripeClient,
		SMTPDialer:                        smtpDialer,
		SMTPSender:                        os.Getenv("SMTP_SENDER"),
		StripeCheckoutSessionSuccessURL:   os.Getenv("STRIPE_CHECKOUT_SESSION_SUCCESS_URL"),
		StripeCheckoutSessionCancelURL:    os.Getenv("STRIPE_CHECKOUT_SESSION_CANCEL_URL"),
		StripeCheckoutSessionPriceID:      os.Getenv("STRIPE_CHECKOUT_SESSION_PRICE_ID"),
		StripeWebhookSigningSecret:        os.Getenv("STRIPE_WEBHOOK_SIGNING_SECRET"),
		PublicURLScheme:                   os.Getenv("PUBLIC_URL_SCHEME"),
		AuthgearOnceDownloadURLGoTemplate: authgearOnceDownloadURLGoTemplate,
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
