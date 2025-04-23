package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-once-license-server/pkg/httpmiddleware"
)

var rootCmd = &cobra.Command{
	Use: "authgear-once-license-server",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server at port 8200",
	Run: func(cmd *cobra.Command, args []string) {
		mux := http.NewServeMux()
		cors := httpmiddleware.CORSMiddleware(os.Getenv("CORS_ALLOWED_ORIGINS"))

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "it works")
		})

		err := http.ListenAndServe(":8200", cors(mux))
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
