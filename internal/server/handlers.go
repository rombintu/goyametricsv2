package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/rombintu/goyametricsv2/internal/logger"
	models "github.com/rombintu/goyametricsv2/internal/models"
	"github.com/rombintu/goyametricsv2/internal/storage"
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
	// Таким макаром можно сериализовать запрос
	// if err := c.Bind(&metric); err != nil {
	// 	return c.String(http.StatusBadRequest, err.Error())
	// }

	// А таким отправлять
	// return c.JSON(http.StatusOK, metric)

	if err := json.NewDecoder(c.Request().Body).Decode(&metric); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	logger.Log.Debug(
		"Try decode metric",
		zap.String("id", metric.ID),
		zap.String("mtype", metric.MType),
		// Далее код не работает, тк есть omitempty
		// zap.Int64("delta", *metric.Delta),
		// zap.Float64("value", *metric.Value),
	)

	// Создаем ошибку для обработки пустых значения
	err := errors.New("delta or value must be not null")

	var mvalue string
	// Парсим то что нужно, взависимости от типа, делаем строку чтобы не менять логику
	switch metric.MType {
	case storage.GaugeType:
		if metric.Value == nil {
			logger.Log.Error(err.Error())
			return c.String(http.StatusBadRequest, err.Error())
		}
		// Если все ок то парсим
		mvalue = strconv.FormatFloat(*metric.Value, 'g', -1, 64)
	case storage.CounterType:
		if metric.Delta == nil {
			logger.Log.Error(err.Error())
			return c.String(http.StatusBadRequest, err.Error())
		}
		// Если все ок то парсим
		mvalue = strconv.FormatInt(*metric.Delta, 10)
	}

	logger.Log.Debug("Parse", zap.String("value", mvalue))

	if err := s.storage.Update(metric.MType, metric.ID, mvalue); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Если 0 то синхронная запись. По хорошему это засунуть в мидлварю конечно
	if s.config.SyncMode {
		s.syncStorage()
	}

	// Костыли для тз
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return json.NewEncoder(c.Response()).Encode(metric)
}

// route for /value. Content-Type: application/json
func (s *Server) MetricValueHandlerJSON(c echo.Context) error {
	var metric models.Metrics

	if err := json.NewDecoder(c.Request().Body).Decode(&metric); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	mvalue, err := s.storage.Get(metric.MType, metric.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "not found")
	}

	// Функция чтобы не повторяться в агенте
	if err := metric.SetValueOrDelta(mvalue); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
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
