.PHONY: mockgen test lint
mockgen:
	mockgen -source=pkg/openfeature/provider.go -destination=pkg/openfeature/provider_mock_test.go -package=openfeature
	mockgen -source=pkg/openfeature/hooks.go -destination=pkg/openfeature/hooks_mock_test.go -package=openfeature
	mockgen -source=pkg/openfeature/mutex.go -destination=pkg/openfeature/mutex_mock_test.go -package=openfeature
test:
	go test -cover ./...
lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	${GOPATH}/bin/golangci-lint run --deadline=3m --timeout=3m ./... # Run linters
