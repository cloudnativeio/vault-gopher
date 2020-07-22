package main

import (
	"github.com/trx35479/vault-gopher/pkg/apis"
	"github.com/trx35479/vault-gopher/pkg/log"
)

func main() {
	var logger = log.NewLogger()
	logger.Println("Secret has been created...")
	c := apis.Client{}
	c.GetClientToken(sfsdf)
}
