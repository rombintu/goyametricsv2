package logger

import (
	"time"

	"github.com/labstack/echo"
	"go.uber.org/zap"
)

const (
	debugLevel = "debug"
	infoLevel  = "info"

	DevMode  = "dev"
	ProdMode = "prod"
)

// Взял пример из урока, реализация логгера по паттерну Singleton
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(mode string) (err error) {
	var cfg zap.Config
	var lvl zap.AtomicLevel
	switch mode {
	case ProdMode:
		cfg = zap.NewProductionConfig()
		lvl, err = zap.ParseAtomicLevel(infoLevel)
		if err != nil {
			return err
		}
	default:
		cfg = zap.NewDevelopmentConfig()
		lvl, err = zap.ParseAtomicLevel(debugLevel)
		if err != nil {
			return err
		}
	}

	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	defer zl.Sync()
	// устанавливаем синглтон
	Log = zl
	return nil
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
// Custom middleware
func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		startTimestamp := time.Now()
		err := next(c)
		if err != nil {
			c.Error(err)
		}
		req := c.Request()
		res := c.Response()

		duration := time.Since(startTimestamp)
		Log.Info("REQEST",
			zap.String("URI", req.URL.Path),
			zap.String("Method", req.Method),
			zap.String("Duration", duration.String()),
			zap.String("Content-Type", req.Header.Get("Content-Type")),
		)
		Log.Info("RESPONSE",
			zap.Int("Status Code", res.Status),
			zap.Int64("Size", res.Size),
			zap.String("Content-Type", res.Header().Get("Content-Type")),
		)
		return err
	}
}
