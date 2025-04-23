package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
	"gopkg.in/gomail.v2"

	"github.com/authgear/authgear-once-license-server/pkg/emailtemplate"
	"github.com/authgear/authgear-once-license-server/pkg/httpmiddleware"
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

var rootCmd = &cobra.Command{
	Use: "authgear-once-license-server",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server at port 8200",
	Run: func(cmd *cobra.Command, args []string) {
		mux := http.NewServeMux()
		cors := httpmiddleware.CORSMiddleware(os.Getenv("CORS_ALLOWED_ORIGINS"))
		maxbytes := httpmiddleware.MaxBytesMiddleware(100 * 1000) // 100KB

		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(indexHTML))
		})

		mux.HandleFunc("/v1/stripe/checkout", func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()

			ctx := r.Context()
			deps := GetDependencies(ctx)
			stripeClient := deps.StripeClient

			checkoutSession, err := pkgstripe.NewCheckoutSession(ctx, stripeClient, &pkgstripe.CheckoutSessionParams{
				SuccessURL: deps.StripeCheckoutSessionSuccessURL,
				CancelURL:  deps.StripeCheckoutSessionCancelURL,
				PriceID:    deps.StripeCheckoutSessionPriceID,
			})
			if err != nil {
				log.Printf("failed to create checkout session: %v", err)
				http.Error(w, "failed to create checkout session", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, checkoutSession.URL, http.StatusSeeOther)
		})

		// The API version must be 2025-03-31.basil
		mux.HandleFunc("/v1/stripe/webhook", func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()

			ctx := r.Context()
			deps := GetDependencies(ctx)

			e, err := pkgstripe.ConstructEvent(r, deps.StripeWebhookSigningSecret)
			if err != nil {
				log.Printf("failed to construct webhook event: %v", err)
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
				id, ok := pkgstripe.GetCheckoutSessionID(e)
				if !ok {
					log.Printf("checkout session ID not found: %v", string(b))
					return
				}

				email, ok := pkgstripe.GetCustomerEmail(e)
				if !ok {
					log.Printf("customer email not found: %v", string(b))
					return
				}

				htmlBody := emailtemplate.RenderInstallationEmail(emailtemplate.InstallationEmailData{
					InstallationOneliner: "TODO",
				})

				opts := smtp.EmailOptions{
					Sender:   deps.SMTPSender,
					Subject:  "Installing Authgear once",
					HTMLBody: htmlBody,
					To:       email,
				}

				err = smtp.SendEmail(deps.SMTPDialer, opts)
				if err != nil {
					log.Printf("failed to send email: %v", err)
				}
				log.Printf("sent installation to checkout session %v", id)
			}
		})

		server := &http.Server{
			Addr:    ":8200",
			Handler: maxbytes(cors(mux)),
			BaseContext: func(_ net.Listener) context.Context {
				return cmd.Context()
			},
		}

		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

type dependenciesKeyType struct{}

var dependenciesKey = dependenciesKeyType{}

type Dependencies struct {
	StripeClient                    *client.API
	SMTPDialer                      *gomail.Dialer
	SMTPSender                      string
	StripeCheckoutSessionSuccessURL string
	StripeCheckoutSessionCancelURL  string
	StripeCheckoutSessionPriceID    string
	StripeWebhookSigningSecret      string
}

func GetDependencies(ctx context.Context) Dependencies {
	return ctx.Value(dependenciesKey).(Dependencies)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

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

	dependencies := Dependencies{
		StripeClient:                    stripeClient,
		SMTPDialer:                      smtpDialer,
		SMTPSender:                      os.Getenv("SMTP_SENDER"),
		StripeCheckoutSessionSuccessURL: os.Getenv("STRIPE_CHECKOUT_SESSION_SUCCESS_URL"),
		StripeCheckoutSessionCancelURL:  os.Getenv("STRIPE_CHECKOUT_SESSION_CANCEL_URL"),
		StripeCheckoutSessionPriceID:    os.Getenv("STRIPE_CHECKOUT_SESSION_PRICE_ID"),
		StripeWebhookSigningSecret:      os.Getenv("STRIPE_WEBHOOK_SIGNING_SECRET"),
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, dependenciesKey, dependencies)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
