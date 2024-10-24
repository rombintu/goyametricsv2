package agent

import (
	"bytes"
	"compress/gzip"
	"context"
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
	"github.com/rombintu/goyametricsv2/lib/mygzip"
	"github.com/rombintu/goyametricsv2/lib/myhash"
	"github.com/rombintu/goyametricsv2/lib/patterns"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

type Agent struct {
	serverAddress  string
	pollInterval   int64
	reportInterval int64
	data           Data // TODO
	pollCount      int
	hashKey        string
	rateLimit      int64
	semaphore      *patterns.Semaphore
}

type Data struct {
	Counters []Counter
	Gauges   []Gauge
}

type Counter struct {
	name  string
	value int64
}
type Gauge struct {
	name  string
	value float64
}

func NewAgent(c config.AgentConfig) *Agent {
	return &Agent{
		serverAddress:  fixServerURL(c.Address),
		pollInterval:   c.PollInterval,
		reportInterval: c.ReportInterval,
		data:           Data{},
		hashKey:        c.HashKey,
		rateLimit:      c.RateLimit,
	}
}

func (a *Agent) Configure() {
	// Configure semaphore
	if a.rateLimit > 0 {
		a.semaphore = patterns.NewSemaphore(a.rateLimit)
	}
}

func fixServerURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		return url
	} else {
		return fmt.Sprintf("http://%s", url)
	}
}

func (a *Agent) incPollCount() {
	a.pollCount++
}

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

	req, err := http.NewRequest(http.MethodPost, url, &buff)
	if err != nil {
		return err
	}

	// If secret key is set
	if a.hashKey != "" {
		hashPayload := myhash.ToSHA256AndHMAC(jsonData, a.hashKey)
		req.Header.Set(myhash.Sha256Header, hashPayload)
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
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

// Add one more gorutine
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

	// get random float64 value
	randomValue := rand.Float64()
	gauges = append(gauges, Gauge{
		name:  "RandomValue",
		value: randomValue,
	})
	a.incPollCount()

	a.data.Counters = counters
	a.data.Gauges = gauges

}

func (a *Agent) Ping() error {
	resp, err := http.Get(a.serverAddress + "/ping")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

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
