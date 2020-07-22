package main

import (
	"github.com/trx35479/vault-gopher/secret-injector/apis"
	"github.com/trx35479/vault-gopher/secret-injector/log"
)

func main() {
	var logger = log.NewLogger()
	logger.Println("Secret has been created...")
	c := apis.Client{}
	c.GetClientToken(sfsdf)
}
