package config

import (
	"reflect"
	"testing"
)

func TestLoadServerConfig(t *testing.T) {
	tests := []struct {
		name string
		want ServerConfig
	}{
		{
			name: "try simple load",
			want: ServerConfig{
				Listen:        "localhost:8080",
				StorageDriver: "mem",
				StoreInterval: 300,
				StoragePath:   "store.json",
				RestoreFlag:   true,
				SyncMode:      false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadServerConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadServerConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
