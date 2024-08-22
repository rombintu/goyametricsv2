package server

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/mygzip"
	"go.uber.org/zap"
)

type Server struct {
	config  config.ServerConfig
	storage storage.Storage
	router  *echo.Echo
}

func NewServer(storage storage.Storage, config config.ServerConfig) *Server {
	return &Server{
		config:  config,
		router:  echo.New(),
		storage: storage,
	}
}

func (s *Server) Configure() {
	s.ConfigureRenderer()
	s.ConfigureMiddlewares()
	s.ConfigureRouter()
	s.ConfigureStorage()
}

func (s *Server) Run() {
	logger.Log.Info("Server is starting on: ", zap.String("url", s.config.Listen))
	if err := http.ListenAndServe(s.config.Listen, s.router); err != nil {
		logger.Log.Fatal("cannot run server", zap.Error(err))
	}
}

func (s *Server) ConfigureStorage() {
	if err := s.storage.Open(); err != nil {
		logger.Log.Error("cannot open storage", zap.Error(err))
	}
	// If restore flag is True, then restore the storage
	if s.config.RestoreFlag {
		if err := s.storage.Restore(); err != nil {
			logger.Log.Warn("cannot restore storage", zap.String("error", err.Error()))
		}
	}
}

func (s *Server) ConfigureRouter() {
	s.router.GET("/", s.RootHandler)
	s.router.GET("/value/:mtype/:mname", s.MetricGetHandler)
	s.router.POST("/update/:mtype/:mname/:mvalue", s.MetricsHandler)

	// JSON
	s.router.POST("/update/", s.MetricUpdateHandlerJSON)
	s.router.POST("/value/", s.MetricValueHandlerJSON)
}

func (s *Server) ConfigureMiddlewares() {
	logger.Initialize(s.config.EnvMode)
	s.router.Use(logger.RequestLogger)

	// Реализация gzip middleware в пару строк, больше ничего не нужно
	// s.router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
	// 	Level: middleware.DefaultGzipConfig.Level,
	// }))

	// Реализация gzip middleware для тз
	s.router.Use(mygzip.GzipMiddleware)
}

func (s *Server) syncStorage() {
	if err := s.storage.Save(); err != nil {
		logger.Log.Error("cannot save storage", zap.Error(err))
	}
	logger.Log.Debug("Storage synchronized", zap.String("path", s.config.StorePath))
}

func (s *Server) SyncStorageInterval() {
	s.syncStorage()
}

func (s *Server) Shutdown() {
	logger.Log.Info("Server is shutting down...")
	s.syncStorage()
	s.storage.Close()
}
