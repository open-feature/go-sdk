.PHONY: mockgen test lint e2e-test
mockgen:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=openfeature/provider.go -destination=openfeature/provider_mock.go -package=openfeature
	mockgen -source=openfeature/hooks.go -destination=openfeature/hooks_mock.go -package=openfeature
	mockgen -source=openfeature/interfaces.go -destination=openfeature/interfaces_mock.go -package=openfeature
	mockgen -source=openfeature/multiprovider/strategies.go -destination=openfeature/multiprovider/strategies_mock.go -package multiprovider
	sed -i.old $$'1s;^;//go:build testtools\\n\\n;' **/*_mock.go
	rm -f **/*_mock.go.old

test:
	go test --short -tags testtools -cover ./...
e2e-test:
	 git submodule update --init --recursive && go test -tags testtools -race -cover ./e2e/...
lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.1
	${GOPATH}/bin/golangci-lint run --build-tags testtools --deadline=3m --timeout=3m ./... # Run linters
