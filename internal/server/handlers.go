package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/logger"
	models "github.com/rombintu/goyametricsv2/internal/models"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/myhash"
	"go.uber.org/zap"
)

// MetricsHandler handles HTTP requests to update metrics in the server's storage system.
// It processes incoming requests to update specific metrics based on the provided parameters
// and stores the updated values in the server's storage system.
//
// Endpoint:
//   - URL: /metrics/:mtype/:mname/:mvalue
//   - Method: POST
//
// Parameters:
//   - mtype: The type of the metric (e.g., "counter", "gauge"). This is a path parameter extracted from the URL.
//   - mname: The name or identifier of the metric. This is a path parameter extracted from the URL.
//   - mvalue: The value to be assigned to the metric. This is a path parameter extracted from the URL.
//
// Request Example:
//
//	POST /metrics/counter/requests/10
//
// Response:
//   - Status: 200 OK
//   - Body: "updated"
//   - Status: 400 Bad Request (if there is an error during the update)
//   - Body: Error message
//   - Status: 404 Not Found (if the metric name is missing)
//   - Body: "Missing metric name"
func (s *Server) MetricsHandler(c echo.Context) error {
	// Extract the metric type, name, and value from the request parameters
	mtype := c.Param("mtype")
	mname := c.Param("mname")
	mvalue := c.Param("mvalue")

	// Check if the metric name is missing
	if mname == "" {
		// Return a 404 Not Found status with an error message
		return c.String(http.StatusNotFound, "Missing metric name")
	}
	// Attempt to update the metric in the storage system
	if err := s.storage.Update(mtype, mname, mvalue); err != nil {
		// Log the error with additional context
		logger.Log.Error(
			err.Error(),
			zap.String("type", mtype),
			zap.String("id/name", mname),
			zap.String("value", mvalue),
		)
		// Return a 400 Bad Request status with the error message
		return c.String(http.StatusBadRequest, err.Error())
	}

	// If sync mode is enabled, perform a synchronous storage update
	if s.config.SyncMode {
		s.SyncStorage()
	}
	// Return a 200 OK status with a success message
	return c.String(http.StatusOK, "updated")
}

// MetricGetHandler handles HTTP requests to retrieve the value of a specific metric from the server's storage system.
// It processes incoming requests to fetch the value of a metric based on the provided parameters.
//
// Endpoint:
//   - URL: /metrics/:mtype/:mname
//   - Method: GET
//
// Parameters:
//   - mtype: The type of the metric (e.g., "counter", "gauge"). This is a path parameter extracted from the URL.
//   - mname: The name or identifier of the metric. This is a path parameter extracted from the URL.
//
// Request Example:
//
//	GET /metrics/counter/requests
//
// Response:
//   - Status: 200 OK
//   - Body: The value of the metric
//   - Status: 404 Not Found (if the metric is not found)
//   - Body: "not found"
func (s *Server) MetricGetHandler(c echo.Context) error {
	// Extract the metric type and name from the request parameters
	mtype := c.Param("mtype")
	mname := c.Param("mname")
	// Attempt to retrieve the metric value from the storage system
	value, err := s.storage.Get(mtype, mname)
	if err != nil {
		// Log the error with additional context
		logger.Log.Error(err.Error(), zap.String("type", mtype), zap.String("id/name", mname))
		// Return a 404 Not Found status with an error message
		return c.String(http.StatusNotFound, "not found")
	}
	// Return a 200 OK status with the metric value
	return c.String(http.StatusOK, value)
}

// RootHandler handles HTTP requests to render the root page of the server, displaying all metrics.
// It processes incoming requests to fetch all metrics from the storage system and renders them using a template.
//
// Endpoint:
//   - URL: /
//   - Method: GET
//
// Request Example:
//
//	GET /
//
// Response:
//   - Status: 200 OK
//   - Body: Rendered HTML content displaying all metrics
func (s *Server) RootHandler(c echo.Context) error {
	// Render the metrics.html template with all metrics from the storage system
	return c.Render(http.StatusOK, "metrics.html", s.storage.GetAll())
}

// MetricUpdateHandlerJSON handles HTTP requests to update metrics in the server's storage system using JSON payloads.
// It processes incoming requests to update specific metrics based on the provided JSON payload and stores the updated values in the server's storage system.
//
// Endpoint:
//   - URL: /update
//   - Method: POST
//   - Content-Type: application/json
//
// Request Body Example:
//
//	{
//	  "id": "requests",
//	  "mtype": "counter",
//	  "delta": 10,
//	  "value": null
//	}
//
// Response:
//   - Status: 200 OK
//   - Content-Type: application/json
//   - Body: The updated metric in JSON format
//   - Status: 400 Bad Request (if there is an error during the update or parsing)
//   - Body: Error message
//   - Status: 500 Internal Server Error (if there is an error encoding JSON)
//   - Body: "Failed to encode JSON"
func (s *Server) MetricUpdateHandlerJSON(c echo.Context) error {
	// Define a variable to hold the decoded metric
	var metric models.Metrics

	// Decode the JSON payload from the request body into the metric variable
	if err := json.NewDecoder(c.Request().Body).Decode(&metric); err != nil {
		logger.Log.Error(err.Error())
		// Return a 400 Bad Request status with the error message
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Log the decoded metric for debugging purposes
	logger.Log.Debug(
		"Try decode metric",
		zap.String("id", metric.ID),
		zap.String("mtype", metric.MType),
		zap.Any("delta", metric.Delta),
		zap.Any("value", metric.Value),
	)

	// Initialize a variable to hold the parsed metric value
	var mvalue string
	// Parse the metric value based on its type
	switch metric.MType {
	case storage.GaugeType:
		// Ensure the value is not nil for gauge type
		if metric.Value == nil {
			err := errors.New("delta must be not null")
			logger.Log.Error(err.Error())
			return c.String(http.StatusBadRequest, err.Error())
		}
		// Convert the float value to a string
		mvalue = strconv.FormatFloat(*metric.Value, 'g', -1, 64)
	case storage.CounterType:
		// Ensure the delta is not nil for counter type
		if metric.Delta == nil {
			err := errors.New("delta must be not null")
			logger.Log.Error(err.Error())
			return c.String(http.StatusBadRequest, err.Error())
		}
		// Convert the int64 delta to a string
		mvalue = strconv.FormatInt(*metric.Delta, 10)
	}

	// Log the parsed value for debugging purposes
	logger.Log.Debug("Parse", zap.String("value", mvalue))

	// Attempt to update the metric in the storage system
	if err := s.storage.Update(metric.MType, metric.ID, mvalue); err != nil {
		logger.Log.Error(
			err.Error(), zap.String("type", metric.MType),
			zap.String("id", metric.ID), zap.String("value", mvalue),
		)
		// Return a 400 Bad Request status with the error message
		return c.String(http.StatusBadRequest, err.Error())
	}

	// If sync mode is enabled, perform a synchronous storage update
	if s.config.SyncMode {
		s.SyncStorage()
	}

	// If a hash key is configured, add a SHA256 hash to the response header
	if s.config.HashKey != "" {
		bytesData, err := json.Marshal(metric)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to encode JSON")
		}
		c.Response().Header().Set(myhash.Sha256Header, myhash.ToSHA256AndHMAC(bytesData, s.config.HashKey))
	}
	// Set the response content type to JSON
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// Write the HTTP status code
	c.Response().WriteHeader(http.StatusOK)

	// Encode and return the updated metric in JSON format
	return json.NewEncoder(c.Response()).Encode(metric)
}

// MetricUpdatesHandlerJSON handles requests to update metrics in JSON format.
// It decodes the JSON payload from the request body into a slice of Metrics,
// updates the corresponding values in the storage, and returns the updated metrics in the response.
//
// Example Request:
// POST /update
// Content-Type: application/json
//
// [
//
//	{
//	  "id": "metric1",
//	  "type": "counter",
//	  "delta": 5
//	},
//	{
//	  "id": "metric2",
//	  "type": "gauge",
//	  "value": 10.5
//	}
//
// ]
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
// [
//
//	{
//	  "id": "metric1",
//	  "type": "counter",
//	  "delta": 5
//	},
//	{
//	  "id": "metric2",
//	  "type": "gauge",
//	  "value": 10.5
//	}
//
// ]
func (s *Server) MetricUpdatesHandlerJSON(c echo.Context) error {
	// Define a variable to hold the decoded metrics
	var metrics []models.Metrics

	// Decode the JSON payload from the request body into the metric variable
	if err := json.NewDecoder(c.Request().Body).Decode(&metrics); err != nil {
		// Log the error
		logger.Log.Error(err.Error())
		// Return a 400 Bad Request status with the error message
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Log the decoded metric for debugging purposes
	logger.Log.Debug(
		"Try decode metrics", zap.Int("size", len(metrics)),
	)

	data := storage.Data{
		Counters: make(storage.Counters),
		Gauges:   make(storage.Gauges),
	}
	for _, m := range metrics {
		// Check counters
		if m.Delta != nil && m.Value == nil {
			oldValue, exist := data.Counters[m.ID]
			if exist {
				data.Counters[m.ID] = oldValue + *m.Delta
			} else {
				data.Counters[m.ID] = *m.Delta
			}

		} else if m.Value != nil && m.Delta == nil {
			data.Gauges[m.ID] = *m.Value
		} else {
			return errors.New("delta or value must be not null")
		}
	}

	if err := s.storage.UpdateAll(data); err != nil {
		logger.Log.Error(err.Error())
		return err
	}

	// Если 0 то синхронная запись
	if s.config.SyncMode {
		s.SyncStorage()
	}

	// add HashSHA256 to Header
	if s.config.HashKey != "" {
		bytesData, err := json.Marshal(metrics)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to encode JSON")
		}
		c.Response().Header().Set(myhash.Sha256Header, myhash.ToSHA256AndHMAC(bytesData, s.config.HashKey))
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	return json.NewEncoder(c.Response()).Encode(metrics)
}

// MetricValueHandlerJSON handles requests to retrieve the value of a specific metric in JSON format.
// It decodes the JSON payload from the request body into a Metrics struct,
// retrieves the corresponding value from the storage, and returns the metric with its value in the response.
//
// Example Request:
// POST /value
// Content-Type: application/json
//
//	{
//	  "id": "metric1",
//	  "type": "counter"
//	}
//
// Example Response:
// HTTP/1.1 200 OK
// Content-Type: application/json
//
//	{
//	  "id": "metric1",
//	  "type": "counter",
//	  "delta": 5
//	}
func (s *Server) MetricValueHandlerJSON(c echo.Context) error {
	var metric models.Metrics
	// Decode the JSON payload from the request body into the metric variable
	if err := json.NewDecoder(c.Request().Body).Decode(&metric); err != nil {
		logger.Log.Error(err.Error())
		// Return a 400 Bad Request status with the error message
		return c.String(http.StatusBadRequest, err.Error())
	}
	// Retrieve the metric value from the storage
	mvalue, err := s.storage.Get(metric.MType, metric.ID)
	if err != nil {
		// Log the error with additional details
		logger.Log.Error(err.Error(), zap.String("type", metric.MType), zap.String("id", metric.ID))
		// Return a 404 Not Found status with a custom message
		return c.String(http.StatusNotFound, "not found")
	}

	// Set the value or delta of the metric based on the retrieved value
	if err := metric.SetValueOrDelta(mvalue); err != nil {
		logger.Log.Error(err.Error(), zap.String("value", mvalue))
		// Return a 400 Bad Request status with the error message
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Add HashSHA256 to the response header if a hash key is configured
	if s.config.HashKey != "" {
		bytesData, err := json.Marshal(metric)
		if err != nil {
			// Return a 500 Internal Server Error status with a custom message
			return c.String(http.StatusInternalServerError, "Failed to encode JSON")
		}
		c.Response().Header().Set(myhash.Sha256Header, myhash.ToSHA256AndHMAC(bytesData, s.config.HashKey))
	}
	// Set the response content type to JSON and write the response header
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	// Encode the metric with its value and send it in the response
	return json.NewEncoder(c.Response()).Encode(metric)
}

// PingDatabase handles requests to check the connection to the database.
// It attempts to ping the database and returns a status response based on the result.
//
// Example Request:
// GET /ping
//
// Example Response (Success):
// HTTP/1.1 200 OK
// Content-Type: text/plain
//
// # OK
//
// Example Response (Failure):
// HTTP/1.1 500 Internal Server Error
// Content-Type: text/plain
//
// <error message>
func (s *Server) PingDatabase(c echo.Context) error {
	// Attempt to ping the database
	if err := s.storage.Ping(); err != nil {
		// Return a 500 Internal Server Error status with the error message
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// Return a 200 OK status with a success message
	return c.String(http.StatusOK, "OK")
}
