.PHONY: build install test integration-test clean

## build: compile forge binary into bin/
build:
	go build -ldflags="-X main.version=$$(git describe --tags 2>/dev/null || echo dev)" -o bin/forge ./main.go

## install: install forge to GOPATH/bin
install:
	go install ./...

## test: run unit tests
test:
	go test ./...

## integration-test: generate a project and verify it compiles
integration-test: build
	rm -rf /tmp/forge-integration-test
	mkdir -p /tmp/forge-integration-test
	cd /tmp/forge-integration-test && \
		$$OLDPWD/bin/forge new test-api
	cd /tmp/forge-integration-test/test-api && go build ./...
	@echo "✓ Integration test passed"

## clean: remove build artifacts
clean:
	rm -rf bin/
