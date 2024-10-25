package server

import (
	"testing"

	"github.com/rombintu/goyametricsv2/internal/config"
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
			name: "init server",
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
