.PHONY: start
start:
	go run ./cmd/cli serve

.PHONY: mjml
mjml:
	./scripts/npm/node_modules/.bin/mjml ./pkg/emailtemplate/installation_email.gotemplate.mjml -o ./pkg/emailtemplate/installation_email.gotemplate

.PHONY: build
build:
	go build -o authgear-once-license-server -tags "osusergo netgo static_build timetzdata" ./cmd/cli

.PHONY: fmt
fmt:
	find ./pkg ./cmd -name '*.go' | sort | xargs go tool goimports -w -format-only -local github.com/authgear/authgear-once-license-server

.PHONY: lint
lint:
	go vet ./cmd/... ./pkg/...

.PHONY: govulncheck
govulncheck:
	go tool govulncheck -show traces,version,verbose ./...

.PHONY: test
test:
	go test ./...

.PHONY: check-tidy
check-tidy:
	$(MAKE) fmt
	$(MAKE) mjml
	go mod tidy
	test -z "$(shell git status --porcelain)"
