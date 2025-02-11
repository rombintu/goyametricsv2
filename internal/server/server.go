// Package server internal Server
package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	pb "github.com/rombintu/goyametricsv2/internal/server/proto"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/common"
	"github.com/rombintu/goyametricsv2/lib/mycrypt"
	"github.com/rombintu/goyametricsv2/lib/mygzip"
	"github.com/rombintu/goyametricsv2/lib/myhash"
	"github.com/rombintu/goyametricsv2/lib/myorigin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InternalStorage struct {
	privateKey *rsa.PrivateKey
}

// Server represents the main server struct that holds the configuration, storage, and router.
type Server struct {
	pb.UnimplementedMetricsServer
	config          config.ServerConfig // Configuration for the server
	storage         storage.Storage     // Storage interface for managing data
	router          *echo.Echo          // Echo router for handling HTTP requests
	internalStorage InternalStorage
}

// NewServer creates a new instance of the Server with the provided storage and configuration.
// It initializes the router and sets the configuration and storage for the server.
//
// Parameters:
// - storage: The storage interface to be used by the server.
// - config: The configuration for the server.
//
// Returns:
// - A pointer to the newly created Server instance.
func NewServer(storage storage.Storage, config config.ServerConfig) *Server {
	return &Server{
		config:  config,
		router:  echo.New(),
		storage: storage,
	}
}

// Run starts the server by listening on the configured address and handling incoming requests.
// It logs the server's starting URL and handles any errors that occur during the server's operation.
// If an error occurs, it closes the storage and logs a fatal error.
func (s *Server) Run() {
	logger.Log.Info("Server is starting on: ", zap.String("url", s.config.Listen))
	if err := http.ListenAndServe(s.config.Listen, s.router); err != nil {
		// If an error occurs, close the storage and log a fatal error
		s.storage.Close()
		logger.Log.Fatal("cannot run server", zap.Error(err))
	}
}

// ConfigureStorage initializes the storage by opening it and optionally restoring data if the restore flag is set.
// It logs the storage configuration and any errors that occur during the process.
func (s *Server) ConfigureStorage() {
	if err := s.storage.Open(); err != nil {
		logger.Log.Fatal("cannot open storage", zap.Error(err))
	}
	// If the restore flag is true, restore the storage
	if s.config.RestoreFlag {
		if err := s.storage.Restore(); err != nil {
			logger.Log.Warn("cannot restore storage", zap.String("error", err.Error()))
		}
	}
	logger.Log.Debug("Storage configuration",
		zap.String("driver", s.config.StorageDriver),
		zap.String("path", s.config.StoragePath),
	)
}

// ConfigureRouter sets up the routes for the server's router.
// It defines the endpoints for handling various HTTP requests.
func (s *Server) ConfigureRouter() {
	s.router.GET("/", s.RootHandler)
	s.router.GET("/value/:mtype/:mname", s.MetricGetHandler)
	s.router.POST("/update/:mtype/:mname/:mvalue", s.MetricsHandler)

	// JSON endpoints
	s.router.POST("/update/", s.MetricUpdateHandlerJSON)
	s.router.POST("/value/", s.MetricValueHandlerJSON)

	s.router.POST("/updates/", s.MetricUpdatesHandlerJSON)

	s.router.GET("/ping", s.PingDatabase)
}

// ConfigureMiddlewares sets up the middlewares for the server's router.
// It initializes the logger, adds request logging, gzip compression, and hash checking middlewares.
func (s *Server) ConfigureMiddlewares() {
	logger.Initialize(s.config.EnvMode)

	// iter 21
	if s.config.SecureMode {
		s.router.Use(mycrypt.EncryptMiddleware(s.config.PrivateKeyFile))
	}

	s.router.Use(logger.RequestLogger)

	// Gzip middleware for compression
	s.router.Use(mygzip.GzipMiddleware)

	// Hash check middleware for verifying request integrity
	s.router.Use(myhash.HashCheckMiddleware(s.config.HashKey))

	// iter 24
	logger.Log.Debug("Trusted subnet: ", zap.String("network", s.config.TrustedSubnet))
	if s.config.TrustedSubnet != "" {
		s.router.Use(myorigin.OriginMiddleware(s.config.TrustedSubnet))
	}
}

// ConfigurePprof registers the pprof handlers with the server's router.
// This allows for profiling the server's performance.
func (s *Server) ConfigurePprof() {
	pprof.Register(s.router)
}

func (s *Server) ConfigureCrypto() {
	// Если путь установлен
	if s.config.SecureMode {
		// Если ключ по пути неверный
		if !mycrypt.ValidPrivateKey(s.config.PrivateKeyFile) {
			var err error
			var privateKey *rsa.PrivateKey
			privateKey, err = mycrypt.LoadPrivateKey(s.config.PrivateKeyFile)
			if err != nil {
				logger.Log.Debug("Private key is not valid. Generating new private and public keys...")
				// Делаем новый ключ по этому пути
				privateKey, err = mycrypt.GenRSAKeyPair(s.config.PrivateKeyFile)
				if err != nil {
					logger.Log.Error(err.Error())
				}
			}

			s.internalStorage.privateKey = privateKey
		}
	} else {
		logger.Log.Debug("Private key file not set. Skipping...")
	}
}

// syncStorage synchronizes the storage by saving any pending changes.
// It logs any errors that occur during the save process.
func (s *Server) SyncStorage() {
	if err := s.storage.Ping(); err != nil {
		return
	}
	if err := s.storage.Save(); err != nil {
		logger.Log.Error("cannot save storage", zap.Error(err))
	}
	logger.Log.Debug("Storage synchronized", zap.String("path", s.config.StoragePath))
}

// Shutdown gracefully shuts down the server.
// It logs the shutdown process, synchronizes the storage, and closes the storage.
func (s *Server) Shutdown() {
	logger.Log.Info("Server is shutting down...")
	s.SyncStorage()

	// Close storage pools on shutdown
	if err := s.storage.Close(); err != nil {
		logger.Log.Error("cannot close storage", zap.Error(err))
	}
}

func (s *Server) ConfigureProto() {
	// определяем порт для сервера
	if common.IsPortInUse(s.config.GRPCPort) {
		// Заглушка, используем любой порт (для тестов)
		s.config.GRPCPort = 0
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
	if err != nil {
		logger.Log.Error(err.Error())
	}
	// создаём gRPC-сервер без зарегистрированной службы
	serv := grpc.NewServer()
	// регистрируем сервис
	pb.RegisterMetricsServer(serv, s)

	logger.Log.Info("start gRPC server")
	go func() {
		if err := serv.Serve(listen); err != nil {
			logger.Log.Error("gRPC server failed: ", zap.Error(err))
		}
	}()
}

func (s *Server) GetMetric(ctx context.Context, in *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	value, err := s.storage.Get(in.Mtype, in.Mname)
	if err != nil {
		logger.Log.Error(err.Error(), zap.String("type", in.Mtype), zap.String("id/name", in.Mname))
		return nil, status.Errorf(codes.NotFound, "not found: %s - %s", in.Mtype, in.Mname)
	}
	return &pb.GetMetricResponse{
		Metric: &pb.Metric{
			Mtype:  in.Mtype,
			Mname:  in.Mname,
			Mvalue: value,
		},
	}, nil
}

func (s *Server) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	var r pb.UpdateMetricResponse
	if err := s.storage.Update(
		in.Metric.Mtype,
		in.Metric.Mname,
		in.Metric.Mvalue,
	); err != nil {
		logger.Log.Error(err.Error(), zap.String("type", in.Metric.Mtype), zap.String("id/name", in.Metric.Mname))
		return nil, status.Error(codes.Aborted, err.Error())
	}
	r.Metric = in.Metric
	return &r, nil
}
