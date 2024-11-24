// Package agent Agent
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	models "github.com/rombintu/goyametricsv2/internal/models"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/mycrypt"
	"github.com/rombintu/goyametricsv2/lib/mygzip"
	"github.com/rombintu/goyametricsv2/lib/myhash"
	"github.com/rombintu/goyametricsv2/lib/patterns"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

// Agent represents the agent that collects and reports metrics to the server.
type Agent struct {
	serverAddress  string              // The address of the server to which metrics are reported
	pollInterval   int64               // The interval at which metrics are polled
	reportInterval int64               // The interval at which metrics are reported to the server
	data           Data                // The collected metrics data
	pollCount      int                 // The count of polls performed
	hashKey        string              // The key used for hashing the metrics data
	rateLimit      int64               // The rate limit for sending metrics
	semaphore      *patterns.Semaphore // Semaphore to control the rate limit

	// iter 21. Cryptographic
	publicKey     *rsa.PublicKey
	publicKeyFile string
	secureMode    bool
}

// Data represents the collected metrics data, including counters and gauges.
type Data struct {
	Counters []Counter // The collected counter metrics
	Gauges   []Gauge   // The collected gauge metrics
}

// Counter represents a counter metric with a name and value.
type Counter struct {
	name  string // The name of the counter metric
	value int64  // The value of the counter metric
}

// Gauge represents a gauge metric with a name and value.
type Gauge struct {
	name  string  // The name of the gauge metric
	value float64 // The value of the gauge metric
}

// NewAgent creates a new instance of the Agent with the provided configuration.
// It initializes the agent with the server address, poll interval, report interval, and other settings.
//
// Parameters:
// - c: The configuration for the agent.
//
// Returns:
// - A pointer to the newly created Agent instance.
func NewAgent(c config.AgentConfig) *Agent {
	return &Agent{
		serverAddress:  fixServerURL(c.Address),
		pollInterval:   c.PollInterval,
		reportInterval: c.ReportInterval,
		data:           Data{},
		hashKey:        c.HashKey,
		rateLimit:      c.RateLimit,
		secureMode:     c.SecureMode,
		publicKeyFile:  c.PublicKeyFile,
	}
}

// Configure configures the agent by setting up the semaphore if a rate limit is specified.
func (a *Agent) Configure() {
	// Configure semaphore if rate limit is greater than 0
	if a.rateLimit > 0 {
		a.semaphore = patterns.NewSemaphore(a.rateLimit)
	}

	if a.secureMode {
		publicKey, err := mycrypt.LoadPublicKey(a.publicKeyFile)
		if err != nil {
			logger.Log.Error("Failed to load public key", zap.Error(err))
		}
		a.publicKey = publicKey
	}
}

// fixServerURL ensures that the server URL starts with "http://".
// If the URL does not start with "http://", it prepends "http://" to the URL.
//
// Parameters:
// - url: The server URL to be fixed.
//
// Returns:
// - The fixed server URL.
func fixServerURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		return url
	} else {
		return fmt.Sprintf("http://%s", url)
	}
}

// incPollCount increments the poll count by 1.
func (a *Agent) incPollCount() {
	a.pollCount++
}

// postRequestJSON sends a POST request with JSON data to the specified URL.
// It compresses the data using gzip and includes a hash if a secret key is set.
//
// Parameters:
// - url: The URL to which the request is sent.
// - data: The data to be sent in the request body.
//
// Returns:
// - An error if the request fails, otherwise nil.
func (a *Agent) postRequestJSON(url string, data any) error {
	if err := a.TryConnectToServer(); err != nil {
		return err
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Start gzip compression
	var buff bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&buff, gzip.BestCompression)
	if err != nil {
		logger.Log.Error("failed init compress writer", zap.Error(err))
		return err
	}
	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		logger.Log.Error("failed write data to compress temporary buffer", zap.Error(err))
		return err
	}
	gzipWriter.Close()

	// End gzip compression

	// Start crypto
	if err := mycrypt.EncryptWithPublicKey(a.publicKey, &buff); err != nil {
		logger.Log.Error("failed encrypt data with public key", zap.Error(err))
		return err
	}
	// End crypto

	req, err := http.NewRequest(http.MethodPost, url, &buff)
	if err != nil {
		return err
	}

	// If secret key is set, include the hash in the request header
	if a.hashKey != "" {
		hashPayload := myhash.ToSHA256AndHMAC(jsonData, a.hashKey)
		req.Header.Set(myhash.Sha256Header, hashPayload)
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// Set header for gzip compression
	req.Header.Set(echo.HeaderContentEncoding, mygzip.GzipHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// sendAllDataOnServer sends all collected metrics data to the server.
// It converts the data into the appropriate format and sends it using a POST request.
//
// Parameters:
// - data: The data to be sent to the server.
//
// Returns:
// - An error if the request fails, otherwise nil.
func (a *Agent) sendAllDataOnServer(data Data) error {
	url := fmt.Sprintf("%s/updates/", a.serverAddress)
	var metrics []models.Metrics

	for _, c := range data.Counters {
		m := models.Metrics{
			ID:    c.name,
			MType: storage.CounterType,
			Delta: &c.value,
		}
		metrics = append(metrics, m)
	}

	for _, g := range data.Gauges {
		m := models.Metrics{
			ID:    g.name,
			MType: storage.GaugeType,
			Value: &g.value,
		}
		metrics = append(metrics, m)
	}

	if err := a.postRequestJSON(url, metrics); err != nil {
		return err
	}
	return nil
}

// RunReport runs the report worker that sends collected metrics to the server at the specified interval.
// It listens for the context to be done to gracefully shut down.
//
// Parameters:
// - ctx: The context to manage the lifecycle of the worker.
// - wg: The wait group to synchronize the shutdown of the worker.
func (a *Agent) RunReport(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("worker is shutdown", zap.String("name", "report"))
			return
		default:
			a.data.Counters = append(a.data.Counters, Counter{
				name:  "PollCount",
				value: int64(a.pollCount),
			})
			if a.rateLimit > 0 {
				logger.Log.Debug("Acquire", zap.String("worker", "pollv1"))
				a.semaphore.Acquire()
			}
			if err := a.sendAllDataOnServer(a.data); err != nil {
				logger.Log.Debug("message from worker", zap.String("name", "report"), zap.String("error", err.Error()))
				time.Sleep(time.Duration(a.reportInterval) * time.Second)
			}
			if a.rateLimit > 0 {
				logger.Log.Debug("Release", zap.String("worker", "pollv1"))
				a.semaphore.Release()
			}
			time.Sleep(time.Duration(a.reportInterval) * time.Second)
		}
	}
}

// RunPoll runs the poll worker that collects metrics at the specified interval.
// It listens for the context to be done to gracefully shut down.
//
// Parameters:
// - ctx: The context to manage the lifecycle of the worker.
// - wg: The wait group to synchronize the shutdown of the worker.
func (a *Agent) RunPoll(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("worker is shutdown", zap.String("name", "poll"))
			return
		default:
			a.loadMetrics()
			logger.Log.Debug("message from worker", zap.String("name", "poll"), zap.String("action", "load metrics common"))
			time.Sleep(time.Duration(a.pollInterval) * time.Second)
		}
	}
}

// RunPollv2 runs an additional poll worker that collects optional metrics at the specified interval.
// It listens for the context to be done to gracefully shut down.
//
// Parameters:
// - ctx: The context to manage the lifecycle of the worker.
// - wg: The wait group to synchronize the shutdown of the worker.
func (a *Agent) RunPollv2(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("worker is shutdown", zap.String("name", "pollv2"))
			return
		default:
			optData := a.loadPSUtilsMetrics()
			if a.rateLimit > 0 {
				logger.Log.Debug("Acquire", zap.String("worker", "pollv2"))
				a.semaphore.Acquire()
			}
			if err := a.sendAllDataOnServer(optData); err != nil {
				logger.Log.Warn(err.Error())
			}
			if a.rateLimit > 0 {
				logger.Log.Debug("Release", zap.String("worker", "pollv2"))
				a.semaphore.Release()
			}
			logger.Log.Debug("message from worker", zap.String("name", "poll"), zap.String("action", "load metrics optionally"))
			time.Sleep(time.Duration(a.pollInterval) * time.Second)
		}
	}
}

// loadPSUtilsMetrics collects optional metrics using the gopsutil library.
// It collects metrics related to memory and CPU utilization.
//
// Returns:
// - The collected optional metrics data.
func (a *Agent) loadPSUtilsMetrics() Data {
	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Warn(err.Error())
		return Data{}
	}

	u, err := cpu.Percent(0, false)
	if err != nil {
		logger.Log.Warn(err.Error())
		return Data{}
	}

	var newGauges []Gauge
	newGauges = append(newGauges, Gauge{name: "TotalMemory", value: float64(v.Total)})
	newGauges = append(newGauges, Gauge{name: "FreeMemory", value: float64(v.Free)})
	newGauges = append(newGauges, Gauge{name: "CPUutilization1", value: u[0]})

	return Data{
		Gauges: newGauges,
	}
}

// loadMetrics collects common metrics using the runtime package.
// It collects metrics related to memory and CPU utilization.
func (a *Agent) loadMetrics() {
	var metrics runtime.MemStats
	runtime.ReadMemStats(&metrics)

	var metricsInterface map[string]interface{}
	inrec, err := json.Marshal(metrics)
	if err != nil {
		return
	}
	json.Unmarshal(inrec, &metricsInterface)

	var counters []Counter
	var gauges []Gauge
	for name, value := range metricsInterface {
		switch v := value.(type) {
		case float64:
			gauges = append(gauges, Gauge{
				name:  name,
				value: v,
			})
		case int64:
			counters = append(counters, Counter{
				name:  name,
				value: v,
			})
		}
	}

	// Get random float64 value
	randomValue := rand.Float64()
	gauges = append(gauges, Gauge{
		name:  "RandomValue",
		value: randomValue,
	})
	a.incPollCount()

	a.data.Counters = counters
	a.data.Gauges = gauges
}

// Ping sends a GET request to the server's ping endpoint to check the connection.
//
// Returns:
// - An error if the request fails, otherwise nil.
func (a *Agent) Ping() error {
	resp, err := http.Get(a.serverAddress + "/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// TryConnectToServer attempts to connect to the server by sending a ping request.
// It retries the connection up to 5 times with increasing intervals if the initial attempt fails.
//
// Returns:
// - An error if the connection fails after all retries, otherwise nil.
func (a *Agent) TryConnectToServer() error {
	var err error
	if err = a.Ping(); err != nil {
		for i := 1; i <= 5; i += 2 {
			// Try reconnecting after 2 seconds if connection failed
			logger.Log.Debug("Ping failed, trying to reconnect", zap.Int("attempt", i))
			time.Sleep(time.Duration(i) * time.Second)
			if err := a.Ping(); err == nil {
				return nil
			}
		}
	}
	return err
}
