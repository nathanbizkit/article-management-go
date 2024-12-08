package main

import (
	"github.com/nathanbizkit/article-management-go/server"
	_ "go.uber.org/automaxprocs"
)

func main() {
	server.Start()
}
