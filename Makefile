SOURCE_FILES := $(shell find . -type f -name '*.go')

go-stub-package: $(SOURCE_FILES) go.mod go.sum
	go build -o go-stub-package ./cmd/go-stub-package

.PHONY: test
test:
	go test -v ./...

