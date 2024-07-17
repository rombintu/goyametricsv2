package main

import (
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/server"
	"github.com/rombintu/goyametricsv2/internal/storage"
)

func main() {
	config := config.LoadServerConfig()
	storage := storage.NewStorage(config.StorageDriver)
	server := server.NewServer(storage, config)
	server.Start()
}
