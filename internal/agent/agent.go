package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	models "github.com/rombintu/goyametricsv2/internal/models"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"go.uber.org/zap"
)

type Agent struct {
	serverAddress  string
	pollInterval   int64
	reportInterval int64
	data           map[string]interface{}
	pollCount      int
	metrics        map[string]string
}

func NewAgent(c config.AgentConfig) *Agent {
	data := make(map[string]interface{})
	data[storage.CounterType] = make(storage.CounterTable)
	data[storage.GaugeType] = make(storage.GaugeTable)
	return &Agent{
		serverAddress:  fixServerURL(c.Address),
		pollInterval:   c.PollInterval,
		reportInterval: c.ReportInterval,
		data:           data,
		metrics:        make(map[string]string),
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

// Deprecated!
// func (a *Agent) postRequest(url string) error {
// 	req, err := http.NewRequest(http.MethodPost, url, nil)
// 	if err != nil {
// 		return err
// 	}
// 	req.Header.Set("Content-Type", "text/plain")
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()
// 	return nil
// }

func (a *Agent) postRequestJSON(url string, metricData models.Metrics) error {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(metricData); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, &buff)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

func (a *Agent) sendDataOnServer(metricType, metricName string, value string) error {
	// Deprecated
	// url := fmt.Sprintf("%s/update/%s/%s/%s", a.serverAddress, metricType, metricName, value)

	// Actually
	url := fmt.Sprintf("%s/update/", a.serverAddress)
	var m models.Metrics
	m.ID = metricName
	m.MType = metricType

	if err := m.SetValueOrDelta(value); err != nil {
		logger.Log.Error(err.Error())
		return err
	}

	if err := a.postRequestJSON(url, m); err != nil {
		logger.Log.Error(err.Error())
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
			logger.Log.Debug("message from worker", zap.String("name", "report"))
			for metricName, value := range a.metrics {
				a.sendDataOnServer(storage.GaugeType, metricName, value)
			}
			a.sendDataOnServer(storage.CounterType, "PollCount", strconv.Itoa(a.pollCount))
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
			logger.Log.Debug("message from worker", zap.String("name", "poll"))
			a.loadMetrics()
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
	for name, value := range metricsInterface {
		switch v := value.(type) {
		case float64:
			a.metrics[name] = strconv.FormatFloat(v, 'g', -1, 64)
		case int64:
			a.metrics[name] = strconv.FormatInt(v, 10)
		}

	}

	// get random float64 value
	randomValue := rand.Float64()
	a.metrics["RandomValue"] = strconv.FormatFloat(randomValue, 'f', -1, 64)
	a.incPollCount()
}
