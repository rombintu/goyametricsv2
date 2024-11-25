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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// main is the entry point of the application.
// It initializes the agent, configures it, and starts the necessary workers.
// The application listens for termination signals to gracefully shut down.
func main() {
	// Create a context with cancel to manage the lifecycle of the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize a wait group to synchronize the shutdown of all workers
	wg := &sync.WaitGroup{}

	// Load the agent configuration
	conf := config.LoadAgentConfig()

	// Create a new agent instance with the loaded configuration
	a := agent.NewAgent(conf)
	// Initialize the logger with the environment mode from the configuration
	logger.Initialize(conf.EnvMode)
	a.Configure()

	logger.Log.Info("Agent starting", zap.String("address", conf.Address))
	logger.OnStartUp(buildVersion, buildDate, buildCommit)

	// Add and start the poll worker
	wg.Add(1)
	go a.RunPoll(ctx, wg)

	// Add and start the report worker
	wg.Add(1)
	go a.RunReport(ctx, wg)

	// Add and start an additional poll worker (version 2)
	wg.Add(1)
	go a.RunPollv2(ctx, wg)

	// Create a channel to capture termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start a goroutine to listen for termination signals
	go func() {
		// Wait for a signal to be received
		sig := <-sigChan
		logger.Log.Info("Received signal", zap.Any("signal", sig))
		// Cancel the context to signal all workers to shut down
		cancel()
	}()

	// Wait for all workers to finish
	wg.Wait()
	logger.Log.Info("All workers have shut down. Exiting program.")
}
