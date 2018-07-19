all: ## generate pd, gateway & swagger json
	protoc --go_out=plugins=grpc:. proto/pay.proto

.PHONY: server
server: ## run API gateway
	go run server/main.go

.PHONY: client
client: ## run golang server
	go run ./server/main.go

help: ## Display help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

