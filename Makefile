GOLANGCI_LINT_VERSION:=v2.8.0
MOCKGEN_VERSION:=v0.6.0

.PHONY: mockgen
mockgen:
	go install go.uber.org/mock/mockgen@${MOCKGEN_VERSION}
	mockgen -destination=interfaces_mock.go -package=openfeature -build_constraint=testtools -mock_names=FeatureProvider=MockProvider go.openfeature.dev/openfeature/v2 clientEvent,evaluationImpl,Hook,FeatureProvider,StateHandler,Tracker

.PHONY: test
test:
	go test --short -tags testtools -cover -timeout 1m ./...

.PHONY: e2e-test
e2e-test:
	git submodule update --init --recursive && go test -tags testtools -race -cover -timeout 1m ./e2e/...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

.PHONY: fix
fix:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./... --fix

.PHONY: docs
docs:
	go doc -http 
