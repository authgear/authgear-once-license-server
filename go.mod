module github.com/authgear/authgear-once-license-server

go 1.24.2

require (
	github.com/getsentry/sentry-go v0.32.0
	github.com/getsentry/sentry-go/slog v0.32.0
	github.com/iawaknahc/originmatcher v0.0.0-20240717084358-ac10088d8800
	github.com/joho/godotenv v1.5.1
	github.com/samber/slog-multi v1.4.0
	github.com/spf13/cobra v1.9.1
	github.com/stripe/stripe-go/v82 v82.0.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/telemetry v0.0.0-20240522233618-39ace7a40ae7 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

tool (
	golang.org/x/tools/cmd/goimports
	golang.org/x/vuln/cmd/govulncheck
)
