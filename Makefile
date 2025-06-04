GOLANGCI_LINT_VERSION:=v2.1.6

.PHONY: mockgen
mockgen:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=openfeature/provider.go -destination=openfeature/provider_mock.go -package=openfeature
	mockgen -source=openfeature/hooks.go -destination=openfeature/hooks_mock.go -package=openfeature
	mockgen -source=openfeature/interfaces.go -destination=openfeature/interfaces_mock.go -package=openfeature
	mockgen -source=openfeature/multiprovider/strategies.go -destination=openfeature/multiprovider/strategies_mock.go -package multiprovider
	sed -i.old $$'1s;^;//go:build testtools\\n\\n;' **/*_mock.go
	rm -f **/*_mock.go.old

.PHONY: test
test:
	go test --short -tags testtools -cover ./...

.PHONY: e2e-test
e2e-test:
	git submodule update --init --recursive && go test -tags testtools -race -cover ./e2e/...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

.PHONY: fix
fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./... --fix

.PHONY: docs
docs:
	go run golang.org/x/pkgsite/cmd/pkgsite@latest -open .
