.PHONY: start
start:
	go run ./cmd/cli serve

.PHONY: mjml
mjml:
	./scripts/npm/node_modules/.bin/mjml ./pkg/emailtemplate/installation_email.gotemplate.mjml -o ./pkg/emailtemplate/installation_email.gotemplate
