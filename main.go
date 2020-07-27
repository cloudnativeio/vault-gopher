package main

import (
	handler "github.com/trx35479/vault-gopher/secret-injector"
	"github.com/trx35479/vault-gopher/secret-injector/log"
)

func main() {
	var logger = log.NewLogger()
	logger.Println("App starting")
	err := handler.CreateObject("secret")
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("Secret has been created")
}
