package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rombintu/goyametricsv2/internal/agent"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	config := config.LoadAgentConfig()
	a := agent.NewAgent(config)

	logger.Initialize(config.EnvMode)
	logger.Log.Info("Agent starting", zap.String("address", config.Address))

	// Add poll worker
	wg.Add(1)
	go a.RunPoll(ctx, wg)

	// Add report worker
	wg.Add(1)
	go a.RunReport(ctx, wg)

	// Канал для перехвата сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Ожидание сигнала
	go func() {
		sig := <-sigChan
		logger.Log.Info("Received signal", zap.Any("sigrnal", sig))
		cancel()
	}()

	// Ожидание завершения всех горутин
	wg.Wait()
	logger.Log.Info("All workers have shut down. Exiting program.")
}
