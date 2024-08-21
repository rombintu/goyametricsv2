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

	"github.com/labstack/echo"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	models "github.com/rombintu/goyametricsv2/internal/models"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/mygzip"
	"go.uber.org/zap"
)

type Agent struct {
	serverAddress  string
	pollInterval   int64
	reportInterval int64
	data           Data // TODO
	pollCount      int
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

func (a *Agent) sendDataOnServer(metricType, metricName string, value string) error {
	url := fmt.Sprintf("%s/update/", a.serverAddress)
	var m models.Metrics
	m.ID = metricName
	m.MType = metricType

	if err := m.SetValueOrDelta(value); err != nil {
		logger.Log.Error(err.Error())
		return err
	}

	if err := a.postRequestJSON(url, m); err != nil {
		return err
	}
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

			if err := a.sendAllDataOnServer(a.data); err != nil {
				logger.Log.Debug("message from worker", zap.String("name", "report"), zap.String("error", err.Error()))
				time.Sleep(time.Duration(a.reportInterval) * time.Second)
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
			logger.Log.Debug("message from worker", zap.String("name", "poll"), zap.String("action", "load metrics"))
			time.Sleep(time.Duration(a.pollInterval) * time.Second)
		}
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
