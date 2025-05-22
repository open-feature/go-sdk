.PHONY: mockgen test lint e2e-test docs
mockgen:
	go install go.uber.org/mock/mockgen@latest
	mockgen -source=openfeature/provider.go -destination=openfeature/provider_mock.go -package=openfeature
	mockgen -source=openfeature/hooks.go -destination=openfeature/hooks_mock.go -package=openfeature
	mockgen -source=openfeature/interfaces.go -destination=openfeature/interfaces_mock.go -package=openfeature
	sed -i.old $$'1s;^;//go:build testtools\\n\\n;' openfeature/*_mock.go
	rm -f openfeature/*_mock.go.old

test:
	go test --short -tags testtools -cover ./...
e2e-test:
	git submodule update --init --recursive && go test -tags testtools -race -cover ./e2e/...
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run --build-tags testtools --timeout=3m ./...
docs:
	go run golang.org/x/pkgsite/cmd/pkgsite@latest -open .
