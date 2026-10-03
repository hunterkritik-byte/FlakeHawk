BINARY := flakehawk

.PHONY: build test lint fmt clean

build:
	mkdir -p bin
	go build -o bin/$(BINARY) ./cmd/flakehawk

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w cmd internal

clean:
	rm -rf bin dist
