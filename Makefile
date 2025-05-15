.PHONY: mockgen test lint e2e-test docs
mockgen:
	mockgen -source=openfeature/provider.go -destination=openfeature/provider_mock_test.go -package=openfeature
	mockgen -source=openfeature/hooks.go -destination=openfeature/hooks_mock_test.go -package=openfeature
	mockgen -source=openfeature/interfaces.go -destination=openfeature/interfaces_mock_test.go -package=openfeature
test:
	go test --short -cover ./...
e2e-test:
	git submodule update --init --recursive && go test -race -cover ./e2e/...
lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.1
	${GOPATH}/bin/golangci-lint run --deadline=3m --timeout=3m ./... # Run linters
docs:
	go run golang.org/x/pkgsite/cmd/pkgsite@latest -open .
