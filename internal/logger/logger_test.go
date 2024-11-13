package logger

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

type MockLogger struct {
	*zap.Logger
	Messages []string
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Messages = append(m.Messages, msg)
}

func TestOnStartUp(t *testing.T) {
	type args struct {
		bversion string
		bdate    string
		bcommit  string
		expected []string
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
				expected: []string{
					"Build version: v0.0.1",
					fmt.Sprintf("Build date: %s", time.Now().Format(time.RFC3339)),
					"Build commit: g23g321111",
				},
			},
		},
		{
			name: "with_nil",
			args: args{
				bversion: "",
				bdate:    time.Now().Format(time.RFC3339),
				bcommit:  "",
				expected: []string{
					"Build version: N/A",
					fmt.Sprintf("Build date: %s", time.Now().Format(time.RFC3339)),
					"Build commit: N/A",
				},
			},
		},
	}
	for _, tt := range tests {
		if err := Initialize("test"); err != nil {
			t.Errorf("Initialize() error = %v", err)
		}
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			Log = mockLogger
			OnStartUp(tt.args.bversion, tt.args.bdate, tt.args.bcommit)
			assert.Equal(t, tt.args.expected, mockLogger.Messages)
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
