# Golang =====================================

GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
GOGET=go get
BINARY_NAME=backend-api
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build ## go test & go build
.PHONY: build
build: ## build go binary
	$(GOBUILD) -o $(BINARY_NAME) -v
.PHONY: test
test: ## go test
	$(GOTEST) -v ./...
.PHONY: clean
clean: ## remove go bainary
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
.PHONY: run
run: ## run go bainary
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
.PHONY: doc
doc: ## exec godoc localhost:6060
	godoc -http=:6060