.PHONY: mockgen test lint e2e-test
mockgen:
	mockgen -source=pkg/openfeature/provider.go -destination=pkg/openfeature/provider_mock_test.go -package=openfeature
	mockgen -source=pkg/openfeature/hooks.go -destination=pkg/openfeature/hooks_mock_test.go -package=openfeature
test:
	go test --short -cover ./...
e2e-test: # dependent on: docker run -p 8013:8013 -v $PWD/test-harness/testing-flags.json:/testing-flags.json ghcr.io/open-feature/flagd-testbed:latest
	go test -cover ./...
	cd test-harness; git restore testing-flags.json; cd .. # reset testing-flags.json

lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	${GOPATH}/bin/golangci-lint run --deadline=3m --timeout=3m ./... # Run linters
