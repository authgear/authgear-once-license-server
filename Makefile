GIT_HASH ?= git-$(shell git rev-parse --short=12 HEAD)

.env:
	@# -o xtrace means set -x
	@# That causes sh to print out the result of the command.
	@# The top level trace is prefixed with `+ `, while the next level trace is prefixed with `++ `.
	@# We are only interested in the top level trace, so we pipe the trace to grep to keep those with `+ ` only.
	@# Finally we use cut to remove the `+ ` prefix.
	sh -o xtrace .env.example 2>&1 | grep '^+ ' | cut -c '3-' | sed -E "s/'//g" > .env

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

.PHONY: build-image
build-image:
	docker build --pull --file ./cmd/cli/Dockerfile --tag quay.io/theauthgear/authgear-once-license-server:$(GIT_HASH) .

.PHONY: push-image
push-image:
	docker push quay.io/theauthgear/authgear-once-license-server:$(GIT_HASH)
