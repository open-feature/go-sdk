.PHONY: mockgen test lint e2e-test
mockgen:
	mockgen -source=openfeature/provider.go -destination=openfeature/mocks/provider.go -package=mocks
	mockgen -source=openfeature/hooks.go -destination=openfeature/mocks/hooks.go -package=mocks
	mockgen -source=openfeature/interfaces.go -destination=openfeature/mocks/interfaces.go -package=mocks
	sed -i.old $$'1s;^;//go:build testtools\\n\\n;' openfeature/mocks/*.go
	rm -f openfeature/mocks/*.go.old

test:
	go test --short -tags testtools -cover ./...
e2e-test:
	 git submodule update --init --recursive && go test -tags testtools -race -cover ./e2e/...
lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.1
	${GOPATH}/bin/golangci-lint run --deadline=3m --timeout=3m ./... # Run linters
