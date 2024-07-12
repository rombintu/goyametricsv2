package main

import (
	"github.com/rombintu/goyametricsv2/internal/server"
	"github.com/rombintu/goyametricsv2/internal/storage"
)

func main() {
	storage := storage.NewStorage("mem")
	server := server.NewServer(storage)
	server.Start()
}
