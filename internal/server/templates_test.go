package server

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/storage"
)

func TestServer_ConfigureRenderer(t *testing.T) {
	type fields struct {
		config  config.ServerConfig
		storage storage.Storage
		router  *echo.Echo
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "configure renderer",
			fields: fields{
				config:  config.ServerConfig{},
				storage: storage.NewTmpDriver(""),
				router:  echo.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// s := &Server{
			// 	config:  tt.fields.config,
			// 	storage: tt.fields.storage,
			// 	router:  tt.fields.router,
			// }
			// // s.ConfigureRenderer()
		})
	}
}
