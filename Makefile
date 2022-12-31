BIN := goquickfix
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: all
all: build

.PHONY: build
build:
	go build -o $(BIN) ./cmd/$(BIN)

.PHONY: install
install:
	go install ./cmd/$(BIN)

.PHONY: test
test: build
	go test -v -race ./...

.PHONY: lint
lint: $(GOBIN)/staticcheck
	go vet ./...
	staticcheck -checks all ./...

$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: clean
clean:
	rm -f $(BIN)
	go clean
