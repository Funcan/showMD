BIN := showmd
CMD := ./cmd/showmd

.PHONY: build test lint tidy clean

build:
	go build -o $(BIN) $(CMD)

test:
	go test ./...

fmt:
	go fmt ./...

lint: fmt vet
	@if command -v staticcheck >/dev/null 2>&1; then staticcheck ./...; fi

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BIN)
