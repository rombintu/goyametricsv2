package logger

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
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

func TestRequestLogger(t *testing.T) {
	// Создаем наблюдателя для логов
	core, logs := observer.New(zapcore.InfoLevel)
	Log = zap.New(core)

	// Создаем экземпляр echo
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Устанавливаем заголовки для запроса
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("HashSHA256", "hash123")

	// Устанавливаем заголовки для ответа
	rec.Header().Set("Content-Type", "application/json")
	rec.Header().Set("Content-Encoding", "gzip")
	rec.Header().Set("HashSHA256", "hash123")

	// Создаем тестовую функцию для обработки запроса
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	// Вызываем middleware с тестовой функцией
	err := RequestLogger(handler)(c)

	// Проверяем, что ошибки нет
	assert.NoError(t, err)

	// Проверяем, что ответ имеет статус 200
	assert.Equal(t, http.StatusOK, rec.Code)

	// Проверяем, что логи были записаны корректно
	allLogs := logs.All()
	assert.Equal(t, 2, len(allLogs))

	// Проверяем логи запроса
	reqLog := allLogs[0]
	assert.Equal(t, "REQEST", reqLog.Message)
	assert.Equal(t, "/test", reqLog.ContextMap()["URI"])
	assert.Equal(t, "GET", reqLog.ContextMap()["Method"])
	assert.NotEmpty(t, reqLog.ContextMap()["Duration"])
	assert.Equal(t, "application/json", reqLog.ContextMap()["Content-Type"])
	assert.Equal(t, "gzip", reqLog.ContextMap()["Accept-Encoding"])
	assert.Equal(t, "hash123", reqLog.ContextMap()["Hash"])

	// Проверяем логи ответа
	resLog := allLogs[1]
	assert.Equal(t, "RESPONSE", resLog.Message)
	assert.Equal(t, int64(http.StatusOK), resLog.ContextMap()["Status Code"])
	assert.Equal(t, int64(4), resLog.ContextMap()["Size"])
	assert.Equal(t, "application/json", resLog.ContextMap()["Content-Type"])
	assert.Equal(t, "gzip", resLog.ContextMap()["Content-Encoding"])
	assert.Equal(t, "hash123", resLog.ContextMap()["Hash"])
}
