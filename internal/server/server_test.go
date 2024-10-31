package server

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/mocks"
	"github.com/rombintu/goyametricsv2/internal/storage"
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
