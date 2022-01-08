.PHONY: test install

go-list-func:
	go build -o go-list-func ./cmd/go-list-func

install:
	go install ./cmd/go-list-func

test:
	go test -v ./...
