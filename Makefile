BINARY := aci-bot
PKG := ./...
BUILD_DIR := bin

.PHONY: all build run test test-race lint fmt tidy clean cover

all: lint test build

build:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY).exe ./cmd/aci-bot

run:
	go run ./cmd/aci-bot

test:
	go test $(PKG)

test-race:
	go test -race $(PKG)

cover:
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -w .

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html
