package main

import (
	"os"
	"vue-golang-payment-app/baskend-api/infrastructure"
)

func main() {
	infrastructure.Router.Run(os.Getenv("API_SERVER_PORT"))
}
