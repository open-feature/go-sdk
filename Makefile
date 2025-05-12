.PHONY: mockgen test lint e2e-test
mockgen:
	mockgen -source=openfeature/provider.go -destination=openfeature/provider_mock_test.go -package=openfeature
	mockgen -source=openfeature/hooks.go -destination=openfeature/hooks_mock_test.go -package=openfeature
	mockgen -source=openfeature/interfaces.go -destination=openfeature/interfaces_mock_test.go -package=openfeature
test:
	go test --short -cover ./...
e2e-test:
	 git submodule update --init --recursive && go test -race -cover ./e2e/...
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run --timeout=3m ./...
