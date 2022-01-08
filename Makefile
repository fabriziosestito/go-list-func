.PHONY: test install ci-tests ci-linter

GOPATH_DIR=`go env GOPATH`

go-list-func:
	go build -o go-list-func ./cmd/go-list-func

install:
	go install ./cmd/go-list-func

test:
	go test -v ./...

ci-linter:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_DIR)/bin v1.30.0
	@$(GOPATH_DIR)/bin/golangci-lint run -v
