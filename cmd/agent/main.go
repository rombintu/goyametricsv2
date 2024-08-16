package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rombintu/goyametricsv2/internal/agent"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}

	config := config.LoadAgentConfig()
	a := agent.NewAgent(config)

	logger.Initialize(config.EnvMode)
	logger.Log.Info("Agent starting", zap.String("address", config.Address))

	var errConn error
	var ok = false
	if errConn = a.Ping(); errConn != nil {
		for i := 1; i <= 5; i += 2 {
			// Try reconnecting after 2 seconds if connection failed
			logger.Log.Debug("Ping failed, trying to reconnect", zap.Int("attempt", i))
			time.Sleep(time.Duration(i) * time.Second)
			if errConn := a.Ping(); errConn == nil {
				ok = true
				break
			}
		}
	}
	if !ok {
		logger.Log.Fatal("Cannot connect to agent", zap.String("error", errConn.Error()))
		return
	}

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
