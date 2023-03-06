.PHONY: mockgen test lint integration-test
mockgen:
	mockgen -source=pkg/openfeature/provider.go -destination=pkg/openfeature/provider_mock_test.go -package=openfeature
	mockgen -source=pkg/openfeature/hooks.go -destination=pkg/openfeature/hooks_mock_test.go -package=openfeature
	mockgen -source=pkg/openfeature/mutex.go -destination=pkg/openfeature/mutex_mock_test.go -package=openfeature
test:
	go test --short -cover ./...
integration-test: # dependent on: flagd start -f file:test-harness/symlink_testing-flags.json
	go test -cover ./integration/...
	cd test-harness; git restore testing-flags.json; cd .. # reset testing-flags.json

lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	${GOPATH}/bin/golangci-lint run --deadline=3m --timeout=3m ./... # Run linters
