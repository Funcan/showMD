BIN := showmd
CMD := ./cmd/showmd

.PHONY: build test lint vet tidy clean

build:
	go build -o $(BIN) $(CMD)

test:
	go test ./...

lint: vet
	@if command -v staticcheck >/dev/null 2>&1; then staticcheck ./...; fi

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BIN)
