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

func (s *Server) MetricsHandler(c echo.Context) error {
	mtype := c.Param("mtype")
	mname := c.Param("mname")
	mvalue := c.Param("mvalue")

	if mname == "" {
		return c.String(http.StatusNotFound, "Missing metric name")
	}
	if err := s.storage.Update(mtype, mname, mvalue); err != nil {
		logger.Log.Error(
			err.Error(),
			zap.String("type", mtype),
			zap.String("id/name", mname),
			zap.String("value", mvalue),
		)
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Если 0 то синхронная запись. По хорошему это засунуть в мидлварю конечно
	if s.config.SyncMode {
		s.syncStorage()
	}

	return c.String(http.StatusOK, "updated")
}

func (s *Server) MetricGetHandler(c echo.Context) error {
	mtype := c.Param("mtype")
	mname := c.Param("mname")
	value, err := s.storage.Get(mtype, mname)
	if err != nil {
		logger.Log.Error(err.Error(), zap.String("type", mtype), zap.String("id/name", mname))
		return c.String(http.StatusNotFound, "not found")
	}
	return c.String(http.StatusOK, value)
}

func (s *Server) RootHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "metrics.html", s.storage.GetAll())
}

// route for /update. Content-Type: application/json
func (s *Server) MetricUpdateHandlerJSON(c echo.Context) error {
	var metric models.Metrics

	if err := json.NewDecoder(c.Request().Body).Decode(&metric); err != nil {
		logger.Log.Error(err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}

	logger.Log.Debug(
		"Try decode metric",
		zap.String("id", metric.ID),
		zap.String("mtype", metric.MType),
		zap.Any("delta", metric.Delta),
		zap.Any("value", metric.Value),
	)

	var mvalue string
	// Парсим то что нужно, взависимости от типа, делаем строку чтобы не менять логику
	switch metric.MType {
	case storage.GaugeType:
		if metric.Value == nil {
			err := errors.New("delta must be not null")
			logger.Log.Error(err.Error())
			return c.String(http.StatusBadRequest, err.Error())
		}
		// Если все ок то парсим
		mvalue = strconv.FormatFloat(*metric.Value, 'g', -1, 64)
	case storage.CounterType:
		if metric.Delta == nil {
			err := errors.New("delta must be not null")
			logger.Log.Error(err.Error())
			return c.String(http.StatusBadRequest, err.Error())
		}
		// Если все ок то парсим
		mvalue = strconv.FormatInt(*metric.Delta, 10)
	}

	logger.Log.Debug("Parse", zap.String("value", mvalue))

	if err := s.storage.Update(metric.MType, metric.ID, mvalue); err != nil {
		logger.Log.Error(
			err.Error(), zap.String("type", metric.MType),
			zap.String("id", metric.ID), zap.String("value", mvalue),
		)
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Если 0 то синхронная запись. По хорошему это засунуть в мидлварю конечно
	if s.config.SyncMode {
		s.syncStorage()
	}

	// add HashSHA256 to Header
	if s.config.HashKey != "" {
		bytesData, err := json.Marshal(metric)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to encode JSON")
		}
		c.Response().Header().Set(myhash.Sha256Header, myhash.ToSHA256AndHMAC(bytesData, s.config.HashKey))
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	return json.NewEncoder(c.Response()).Encode(metric)
}

// route for /updates. Content-Type: application/json
func (s *Server) MetricUpdatesHandlerJSON(c echo.Context) error {
	var metrics []models.Metrics

	if err := json.NewDecoder(c.Request().Body).Decode(&metrics); err != nil {
		logger.Log.Error(err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}

	logger.Log.Debug(
		"Try decode metrics", zap.Int("size", len(metrics)),
	)

	// Нет времени делать тесты)
	if len(metrics) < 11 {
		for _, m := range metrics {
			logger.Log.Debug(
				"metric", zap.String("name", m.ID),
				zap.String("type", m.MType),
				zap.Any("delta", m.Delta),
				zap.Any("value", m.Value),
			)
		}
	}

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

	// Если 0 то синхронная запись. По хорошему это засунуть в мидлварю конечно
	if s.config.SyncMode {
		s.syncStorage()
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

// route for /value. Content-Type: application/json
func (s *Server) MetricValueHandlerJSON(c echo.Context) error {
	var metric models.Metrics

	if err := json.NewDecoder(c.Request().Body).Decode(&metric); err != nil {
		logger.Log.Error(err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	mvalue, err := s.storage.Get(metric.MType, metric.ID)
	if err != nil {
		logger.Log.Error(err.Error(), zap.String("type", metric.MType), zap.String("id", metric.ID))
		return c.String(http.StatusNotFound, "not found")
	}

	// Функция чтобы не повторяться в агенте
	if err := metric.SetValueOrDelta(mvalue); err != nil {
		logger.Log.Error(err.Error(), zap.String("value", mvalue))
		return c.String(http.StatusBadRequest, err.Error())
	}

	// add HashSHA256 to Header
	if s.config.HashKey != "" {
		bytesData, err := json.Marshal(metric)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to encode JSON")
		}
		c.Response().Header().Set(myhash.Sha256Header, myhash.ToSHA256AndHMAC(bytesData, s.config.HashKey))
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	return json.NewEncoder(c.Response()).Encode(metric)
}

// route for /ping. Content-Type: application/json
func (s *Server) PingDatabase(c echo.Context) error {
	if err := s.storage.Ping(); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "OK")
}
