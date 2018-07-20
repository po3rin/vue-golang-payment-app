.PHONY: pay 
pay: ## run payment-service
	go run payment-service/server/server.go
.PHONY: api
api: ## run backend-api
	go run backend-api/main.go
