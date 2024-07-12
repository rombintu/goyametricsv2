package agent

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/storage"
)

type Agent struct {
	serverURL      string
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
		serverURL:      fixServerURL(c.ServerURL),
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

func (a *Agent) postRequest(url string) error {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

func (a *Agent) sendDataOnServer(metricType, metricName string, value string) error {
	url := fmt.Sprintf("%s/update/%s/%s/%s", a.serverURL, metricType, metricName, value)
	if err := a.postRequest(url); err != nil {
		return err
	}
	return nil
}

func (a *Agent) RunPoll() {
	for {
		a.loadMetrics()
		time.Sleep(time.Duration(a.pollInterval) * time.Second)
	}
}

func (a *Agent) RunReport() {
	for {
		for metricName, value := range a.metrics {
			a.sendDataOnServer(storage.GaugeType, metricName, value)
		}
		a.sendDataOnServer(storage.CounterType, "pollCount", strconv.Itoa(a.pollCount))
		time.Sleep(time.Duration(a.reportInterval) * time.Second)
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
		case int:
			a.metrics[name] = strconv.Itoa(v)
		case float64:
			a.metrics[name] = strconv.FormatFloat(v, 'f', -1, 64)
		case uint64:
			a.metrics[name] = strconv.FormatUint(v, 10)
		}

	}

	// get random float64 value
	randomValue := rand.Float64()
	a.metrics["randomValue"] = strconv.FormatFloat(randomValue, 'f', -1, 64)
	a.incPollCount()
}
