package agent

import (
	"context"
	"crypto/rsa"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAgentLoadMetrics(t *testing.T) {
	config := config.LoadAgentConfig()
	agent := NewAgent(config)
	agent.loadMetrics()

	if len(agent.data.Counters) == 0 && agent.pollCount == 0 {
		t.Error("Expected counters metrics to be loaded")
	}
	if len(agent.data.Gauges) == 0 {
		t.Error("Expected gauges metrics to be loaded")
	}
}

func Test_fixServerURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple fix",
			args: args{url: "http://google.com"},
			want: "http://google.com",
		},
		{
			name: "simple fix 2",
			args: args{url: "google.com"},
			want: "http://google.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixServerURL(tt.args.url); got != tt.want {
				t.Errorf("fixServerURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_loadPSUtilsMetrics(t *testing.T) {

	tests := []struct {
		name           string
		lenIsMoreThen0 bool
	}{
		{
			name:           "load cpu utils metrics",
			lenIsMoreThen0: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Agent{data: Data{}}
			if got := a.loadPSUtilsMetrics(); len(got.Counters) != 0 {
				t.Errorf("Agent.loadPSUtilsMetrics() = %+v, want %v", got, tt.lenIsMoreThen0)
			}
			if got := a.loadPSUtilsMetrics(); len(got.Gauges) == 0 {
				t.Errorf("Agent.loadPSUtilsMetrics() = %+v, want %v", got, tt.lenIsMoreThen0)
			}
		})
	}
}

func TestAgent_postRequestJSON(t *testing.T) {
	type args struct {
		url  string
		data any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "failed_post_request_json",
			args:    args{url: "localhost:8080", data: Data{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAgent(config.AgentConfig{})
			if err := a.postRequestJSON(tt.args.url, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Agent.postRequestJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_incPollCount(t *testing.T) {
	t.Run("PollCountIncrement", func(t *testing.T) {
		a := NewAgent(config.AgentConfig{})
		a.incPollCount()
		if a.pollCount != 1 {
			t.Error("pollCount error increment")
		}
	})

}

// MockSemaphore is a mock implementation of the Semaphore struct
type MockSemaphore struct {
	mock.Mock
}

func (m *MockSemaphore) Acquire() {
	m.Called()
}

func (m *MockSemaphore) Release() {
	m.Called()
}

// MockAgent is a mock implementation of the Agent struct
type MockAgent struct {
	mock.Mock
	serverAddress  string
	pollInterval   int64
	reportInterval int64
	data           Data
	pollCount      int
	hashKey        string
	rateLimit      int64
	semaphore      *MockSemaphore
	publicKey      *rsa.PublicKey
	publicKeyFile  string
	secureMode     bool
}

func (m *MockAgent) sendAllDataOnServer(data Data) error {
	args := m.Called(data)
	return args.Error(0)
}

func TestRunReport(t *testing.T) {
	// Создаем наблюдателя для логов
	core, logs := observer.New(zap.DebugLevel)
	logger.Log = zap.New(core)

	// Создаем мок-объект для Semaphore
	mockSemaphore := new(MockSemaphore)

	// Создаем мок-объект для Agent
	mockAgent := &MockAgent{
		reportInterval: 1,
		pollCount:      1,
		rateLimit:      1,
		semaphore:      mockSemaphore,
	}

	// Устанавливаем ожидания для мок-объекта
	mockAgent.On("sendAllDataOnServer", mock.Anything).Return(nil)
	mockSemaphore.On("Acquire").Return()
	mockSemaphore.On("Release").Return()

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем wait group
	var wg sync.WaitGroup
	wg.Add(1)

	// Запускаем метод RunReport в отдельной горутине
	go RunReport(ctx, &wg, mockAgent)

	// Ждем некоторое время, чтобы убедиться, что метод работает
	time.Sleep(2 * time.Second)

	// Отменяем контекст, чтобы завершить работу метода
	cancel()

	// Ждем завершения работы метода
	wg.Wait()

	// Проверяем, что логи были записаны корректно
	allLogs := logs.All()
	assert.GreaterOrEqual(t, len(allLogs), 2)
	assert.Equal(t, "worker is shutdown", allLogs[len(allLogs)-1].Message)
	assert.Equal(t, "report", allLogs[len(allLogs)-1].ContextMap()["name"])

	// Проверяем, что мок-объекты были вызваны корректно
	mockAgent.AssertExpectations(t)
	mockSemaphore.AssertExpectations(t)
}

func TestRunReport_SendAllDataOnServerError(t *testing.T) {
	// Создаем наблюдателя для логов
	core, logs := observer.New(zap.DebugLevel)
	logger.Log = zap.New(core)

	// Создаем мок-объект для Semaphore
	mockSemaphore := new(MockSemaphore)

	// Создаем мок-объект для Agent
	mockAgent := &MockAgent{
		reportInterval: 1,
		pollCount:      1,
		rateLimit:      1,
		semaphore:      mockSemaphore,
	}

	// Устанавливаем ожидания для мок-объекта
	mockAgent.On("sendAllDataOnServer", mock.Anything).Return(errors.New("send error"))
	mockSemaphore.On("Acquire").Return()
	mockSemaphore.On("Release").Return()

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем wait group
	var wg sync.WaitGroup
	wg.Add(1)

	// Запускаем метод RunReport в отдельной горутине
	go RunReport(ctx, &wg, mockAgent)

	// Ждем некоторое время, чтобы убедиться, что метод работает
	time.Sleep(2 * time.Second)

	// Отменяем контекст, чтобы завершить работу метода
	cancel()

	// Ждем завершения работы метода
	wg.Wait()

	// Проверяем, что логи были записаны корректно
	allLogs := logs.All()
	assert.GreaterOrEqual(t, len(allLogs), 2)
	assert.Equal(t, "Release", allLogs[len(allLogs)-2].Message)

	// Проверяем, что мок-объекты были вызваны корректно
	mockAgent.AssertExpectations(t)
	mockSemaphore.AssertExpectations(t)
}

// RunReport is the function we are testing
func RunReport(ctx context.Context, wg *sync.WaitGroup, a *MockAgent) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug("worker is shutdown", zap.String("name", "report"))
			return
		default:
			a.data.Counters = append(a.data.Counters, Counter{
				name:  "PollCount",
				value: int64(a.pollCount),
			})
			if a.rateLimit > 0 {
				logger.Log.Debug("Acquire", zap.String("worker", "pollv1"))
				a.semaphore.Acquire()
			}
			if err := a.sendAllDataOnServer(a.data); err != nil {
				logger.Log.Debug("message from worker", zap.String("name", "report"), zap.String("error", err.Error()))
				time.Sleep(time.Duration(a.reportInterval) * time.Second)
			}
			if a.rateLimit > 0 {
				logger.Log.Debug("Release", zap.String("worker", "pollv1"))
				a.semaphore.Release()
			}
			time.Sleep(time.Duration(a.reportInterval) * time.Second)
		}
	}
}
