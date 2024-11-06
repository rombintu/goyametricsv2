package logger

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestInitialize(t *testing.T) {
	Log := zap.NewNop()
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

func TestOnStartUp(t *testing.T) {
	type args struct {
		bversion string
		bdate    string
		bcommit  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "without_nil",
			args: args{
				bversion: "v0.0.1",
				bdate:    time.Now().Format(time.RFC3339),
				bcommit:  "g23g321111",
			},
		},
		{
			name: "with_nil",
			args: args{
				bversion: "",
				bdate:    time.Now().Format(time.RFC3339),
				bcommit:  "",
			},
		},
	}
	for _, tt := range tests {
		if err := Initialize("test"); err != nil {
			t.Errorf("Initialize() error = %v", err)
		}
		t.Run(tt.name, func(t *testing.T) {
			OnStartUp(tt.args.bversion, tt.args.bdate, tt.args.bcommit)
		})
	}
}

func Test_ifEmptyOpt(t *testing.T) {
	type args struct {
		opt string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "not_empty",
			args: args{
				opt: "some optional argument",
			},
			want: "some optional argument",
		},
		{
			name: "empty",
			args: args{
				opt: "",
			},
			want: "N/A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ifEmptyOpt(tt.args.opt); got != tt.want {
				t.Errorf("ifEmptyOpt() = %v, want %v", got, tt.want)
			}
		})
	}
}
