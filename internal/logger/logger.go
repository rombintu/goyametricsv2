// Package logger
package logger

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	debugLevel = "debug"
	infoLevel  = "info"

	DevMode  = "dev"
	ProdMode = "prod"
)

type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Level() zapcore.Level
	Warn(msg string, fields ...zapcore.Field)
}

// Взял пример из урока, реализация логгера по паттерну Singleton
var Log Logger = zap.NewNop()

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
			// add for Iter7
			zap.String("Content-Type", req.Header.Get("Content-Type")),
			// add for Iter8
			zap.String("Accept-Encoding", req.Header.Get("Accept-Encoding")),
			// add for Iter 14
			zap.String("Hash", req.Header.Get("HashSHA256")),
		)
		Log.Info("RESPONSE",
			zap.Int("Status Code", res.Status),
			zap.Int64("Size", res.Size),
			// add for Iter7
			zap.String("Content-Type", res.Header().Get("Content-Type")),
			// add for Iter8
			zap.String("Content-Encoding", res.Header().Get("Content-Encoding")),
			zap.String("Hash", res.Header().Get("HashSHA256")),
		)
		return err
	}
}

func ifEmptyOpt(opt string) string {
	if opt == "" {
		return "N/A"
	}
	return opt
}

func OnStartUp(bversion, bdate, bcommit string) {
	bversion = ifEmptyOpt(bversion)
	bdate = ifEmptyOpt(bdate)
	bcommit = ifEmptyOpt(bcommit)

	Log.Info(fmt.Sprintf("Build version: %s", bversion))
	Log.Info(fmt.Sprintf("Build date: %s", bdate))
	Log.Info(fmt.Sprintf("Build commit: %s", bcommit))
}
