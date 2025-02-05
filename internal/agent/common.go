package agent

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Ping sends a GET request to the server's ping endpoint to check the connection.
//
// Returns:
// - An error if the request fails, otherwise nil.
func Ping(host, serverAddress string) error {
	req, err := http.NewRequest(http.MethodGet, serverAddress+"/ping", nil)
	if err != nil {
		return err
	}
	req.Header.Set(echo.HeaderXRealIP, host)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// loadPSUtilsMetrics collects optional metrics using the gopsutil library.
// It collects metrics related to memory and CPU utilization.
//
// Returns:
// - The collected optional metrics data.
func loadPSUtilsMetrics() Data {
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
