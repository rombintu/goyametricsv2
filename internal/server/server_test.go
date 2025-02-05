package server

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/mocks"
	pb "github.com/rombintu/goyametricsv2/internal/server/proto"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewServer(t *testing.T) {
	type args struct {
		storage storage.Storage
		config  config.ServerConfig
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "init_server",
			args: args{storage: storage.NewTmpDriver(""), config: config.ServerConfig{}},
			want: &Server{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewServer(tt.args.storage, tt.args.config)
		})
	}
}

func TestSyncStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	t.Run("check_sync_storage", func(t *testing.T) {

		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		// Создаем экземпляр Server с моками
		server := NewServer(
			m,
			config.ServerConfig{},
		)
		defer server.Shutdown()

		// Вызываем метод syncStorage
		server.SyncStorage()
	})
}

func TestConfigureStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	t.Run("Simple_configure_storage", func(t *testing.T) {
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Open().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		server := NewServer(
			m,
			config.ServerConfig{},
		)
		defer server.Shutdown()
		server.ConfigureStorage()
	})
}

func TestConfigureMiddlewares(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	t.Run("Simple_configure_middlewares", func(t *testing.T) {
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		// Создаем экземпляр Server с моками
		server := NewServer(
			m,
			config.ServerConfig{},
		)
		defer server.Shutdown()

		// Вызываем метод
		server.ConfigureMiddlewares()
	})
}

func TestConfigurePprof(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	t.Run("Simple_configure_pprof", func(t *testing.T) {
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		// Создаем экземпляр Server с моками
		server := NewServer(
			m,
			config.ServerConfig{},
		)
		defer server.Shutdown()

		// Вызываем метод
		server.ConfigurePprof()
	})
}

func TestConfigureCrypto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	t.Run("Simple_configure_crypto", func(t *testing.T) {
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		// Создаем экземпляр Server с моками
		server := NewServer(
			m,
			config.ServerConfig{},
		)
		defer server.Shutdown()

		// Вызываем метод
		server.ConfigureCrypto()
	})
}

func TestConfigureProto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	t.Run("Simple_configure_proto", func(t *testing.T) {
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		// Создаем экземпляр Server с моками
		server := NewServer(
			m,
			config.ServerConfig{},
		)
		defer server.Shutdown()

		// Вызываем метод
		server.ConfigureProto()
	})
}

func TestServer_GetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	server := &Server{
		storage: mockStorage,
	}

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &pb.GetMetricRequest{
			Mtype: gaugeMetricType,
			Mname: "alloc",
		}

		mockStorage.EXPECT().Get(gaugeMetricType, "alloc").Return("123.45", nil).Times(1)

		resp, err := server.GetMetric(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "123.45", resp.Metric.Mvalue)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := &pb.GetMetricRequest{
			Mtype: gaugeMetricType,
			Mname: "unknown",
		}

		mockStorage.EXPECT().Get(gaugeMetricType, "unknown").Return("0", errors.New("not found")).Times(1)

		resp, err := server.GetMetric(ctx, req)

		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestServer_UpdateMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)

	server := &Server{
		storage: mockStorage,
	}

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &pb.UpdateMetricRequest{
			Metric: &pb.Metric{
				Mtype:  gaugeMetricType,
				Mname:  "alloc",
				Mvalue: "123.45",
			},
		}

		mockStorage.EXPECT().Update(gaugeMetricType, "alloc", "123.45").Return(nil).Times(1)

		resp, err := server.UpdateMetric(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, req.Metric, resp.Metric)
	})

	t.Run("UpdateError", func(t *testing.T) {
		req := &pb.UpdateMetricRequest{
			Metric: &pb.Metric{
				Mtype:  gaugeMetricType,
				Mname:  "alloc",
				Mvalue: "123.45",
			},
		}

		mockStorage.EXPECT().Update(gaugeMetricType, "alloc", "123.45").Return(errors.New("update failed")).Times(1)

		resp, err := server.UpdateMetric(ctx, req)

		assert.Nil(t, resp)
		assert.Equal(t, codes.Aborted, status.Code(err))
	})
}

func TestServer_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	router := echo.New()
	router.GET("/", echo.HandlerFunc(func(c echo.Context) error {
		return nil
	}))
	server := &Server{
		config:  config.ServerConfig{Listen: "localhost:8080"},
		storage: m,
		router:  router,
	}
	go func() {
		server.Run()
	}()

	// Выполняем тестовый запрос
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	m.EXPECT().Ping().Return(nil).AnyTimes()
	m.EXPECT().Save().Return(nil).AnyTimes()
	m.EXPECT().Open().Return(nil).AnyTimes()
	m.EXPECT().Close().Return(nil).AnyTimes()

	// Останавливаем сервер
	server.storage.Close()
	server.Shutdown()
}
