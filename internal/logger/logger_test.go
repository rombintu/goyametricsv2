package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestInitialize(t *testing.T) {
	var Log *zap.Logger = zap.NewNop()
	type args struct {
		mode string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "init prod",
			args: args{mode: DevMode},
		},
		{
			name: "init prod",
			args: args{mode: ProdMode},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Initialize(tt.args.mode); err != nil {
				t.Errorf("Initialize() error = %v", err)
			}
			Log.Debug("debug message")
		})
	}
}
