package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/internal/server"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"go.uber.org/zap"
)

func main() {
	config := config.LoadServerConfig()
	storage := storage.NewStorage(config.StorageDriver, config.StorePath)
	server := server.NewServer(storage, config)
	server.Configure()
	go server.Run()

	// Не очень разбираюсь в каналах, беру примеры из инета и подгоняю

	// Канал для сигнала завершения
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Duration(config.StoreInterval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				server.SyncStorageInterval()
			case <-done:
				logger.Log.Debug("worker is shutdown", zap.String("name", "sync_storage"))
				return
			}
		}
	}()

	// Канал для перехвата сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Ожидание сигнала
	<-sigChan

	// Ожидание завершения всех горутин
	close(done)
	server.Shutdown()
	logger.Log.Info("All workers have shut down. Exiting program.")

}
