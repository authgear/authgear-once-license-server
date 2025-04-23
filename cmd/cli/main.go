package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v82/client"

	"github.com/authgear/authgear-once-license-server/pkg/httpmiddleware"
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

		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(indexHTML))
		})

		mux.HandleFunc("POST /v1/stripe/checkout", func(w http.ResponseWriter, r *http.Request) {
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
	StripeCheckoutSessionSuccessURL string
	StripeCheckoutSessionCancelURL  string
	StripeCheckoutSessionPriceID    string
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
	dependencies := Dependencies{
		StripeClient:                    stripeClient,
		StripeCheckoutSessionSuccessURL: os.Getenv("STRIPE_CHECKOUT_SESSION_SUCCESS_URL"),
		StripeCheckoutSessionCancelURL:  os.Getenv("STRIPE_CHECKOUT_SESSION_CANCEL_URL"),
		StripeCheckoutSessionPriceID:    os.Getenv("STRIPE_CHECKOUT_SESSION_PRICE_ID"),
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, dependenciesKey, dependencies)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
