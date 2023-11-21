.PHONY: mockgen test lint e2e-test
mockgen:
	mockgen -source=openfeature/provider.go -destination=openfeature/provider_mock_test.go -package=openfeature
	mockgen -source=openfeature/hooks.go -destination=openfeature/hooks_mock_test.go -package=openfeature
test:
	go test --short -cover ./...
e2e-test:
	go test -race -cover ./e2e/...
lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.1
	${GOPATH}/bin/golangci-lint run --deadline=3m --timeout=3m ./... # Run linters
