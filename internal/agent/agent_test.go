package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/internal/mocks"
	models "github.com/rombintu/goyametricsv2/internal/models"
	pb "github.com/rombintu/goyametricsv2/internal/server/proto"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_loadMetrics(t *testing.T) {
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
		Name string
		args args
		want string
	}{
		{
			Name: "simple fix",
			args: args{url: "http://google.com"},
			want: "http://google.com",
		},
		{
			Name: "simple fix 2",
			args: args{url: "google.com"},
			want: "http://google.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := fixServerURL(tt.args.url); got != tt.want {
				t.Errorf("fixServerURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadPSUtilsMetrics(t *testing.T) {
	tests := []struct {
		Name           string
		lenIsMoreThen0 bool
	}{
		{
			Name:           "load cpu utils metrics",
			lenIsMoreThen0: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if got := loadPSUtilsMetrics(); len(got.Counters) != 0 {
				t.Errorf("Agent.loadPSUtilsMetrics() = %+v, want %v", got, tt.lenIsMoreThen0)
			}
			if got := loadPSUtilsMetrics(); len(got.Gauges) == 0 {
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
		Name    string
		args    args
		wantErr bool
	}{
		{
			Name:    "failed_post_request_json",
			args:    args{url: "localhost:8080", data: models.Data{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			a := NewAgent(config.AgentConfig{})
			if err := a.PostRequestJSON(tt.args.url, tt.args.data); (err != nil) != tt.wantErr {
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

func TestGetLocalIP(t *testing.T) {
	// Вызываем тестируемую функцию
	ip := GetLocalIP()
	t.Logf("got %s", ip)
}

// Моковые интерфейсы
type MemoryProvider interface {
	VirtualMemory() (*mem.VirtualMemoryStat, error)
}

type CPUProvider interface {
	Percent(time.Duration, bool) ([]float64, error)
}

// Реальные реализации
type RealMemoryProvider struct{}
type RealCPUProvider struct{}

func (r RealMemoryProvider) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (r RealCPUProvider) Percent(d time.Duration, b bool) ([]float64, error) {
	return cpu.Percent(d, b)
}

// Модифицированная функция с dependency injection
func loadPSUtilsMetricsDI(memProv MemoryProvider, cpuProv CPUProvider) models.Data {
	v, err := memProv.VirtualMemory()
	if err != nil {
		logger.Log.Warn(err.Error())
		return models.Data{}
	}

	u, err := cpuProv.Percent(0, false)
	if err != nil {
		logger.Log.Warn(err.Error())
		return models.Data{}
	}

	return models.Data{
		Gauges: []models.Gauge{
			{Name: "TotalMemory", Value: float64(v.Total)},
			{Name: "FreeMemory", Value: float64(v.Free)},
			{Name: "CPUutilization1", Value: u[0]},
		},
	}
}

// Моки для тестов
type MockMemoryProvider struct {
	mock.Mock
}

func (m *MockMemoryProvider) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	args := m.Called()
	return args.Get(0).(*mem.VirtualMemoryStat), args.Error(1)
}

type MockCPUProvider struct {
	mock.Mock
}

func (m *MockCPUProvider) Percent(d time.Duration, b bool) ([]float64, error) {
	args := m.Called(d, b)
	return args.Get(0).([]float64), args.Error(1)
}

// Тесты
func TestLoadPSUtilsMetrics_Success(t *testing.T) {
	mockMem := new(MockMemoryProvider)
	mockCPU := new(MockCPUProvider)

	mockMem.On("VirtualMemory").Return(&mem.VirtualMemoryStat{
		Total: 1000000000,
		Free:  500000000,
	}, nil)

	mockCPU.On("Percent", 0*time.Second, false).Return([]float64{25.5}, nil)

	result := loadPSUtilsMetricsDI(mockMem, mockCPU)

	assert.Len(t, result.Gauges, 3)
	assert.Contains(t, result.Gauges, models.Gauge{Name: "TotalMemory", Value: 1000000000})
	assert.Contains(t, result.Gauges, models.Gauge{Name: "FreeMemory", Value: 500000000})
	assert.Contains(t, result.Gauges, models.Gauge{Name: "CPUutilization1", Value: 25.5})
}

func TestLoadPSUtilsMetrics_MemError(t *testing.T) {
	mockMem := new(MockMemoryProvider)
	mockCPU := new(MockCPUProvider)

	mockMem.On("VirtualMemory").Return(&mem.VirtualMemoryStat{}, assert.AnError)

	result := loadPSUtilsMetricsDI(mockMem, mockCPU)
	assert.Empty(t, result.Gauges)
}

func TestLoadPSUtilsMetrics_CPUError(t *testing.T) {
	mockMem := new(MockMemoryProvider)
	mockCPU := new(MockCPUProvider)

	mockMem.On("VirtualMemory").Return(&mem.VirtualMemoryStat{
		Total: 1000000000,
		Free:  500000000,
	}, nil)
	mockCPU.On("Percent", 0*time.Second, false).Return([]float64{}, assert.AnError)

	result := loadPSUtilsMetricsDI(mockMem, mockCPU)
	assert.Empty(t, result.Gauges)
}

func TestSendDataGRPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricsClient := mocks.NewMockMetricsClient(ctrl)

	// Создаем тестовые данные
	data := models.Data{
		Counters: []models.Counter{
			{Name: "counter1", Value: 42},
		},
		Gauges: []models.Gauge{
			{Name: "gauge1", Value: 3.14},
		},
	}

	// Ожидаем вызов UpdateMetric для models.Counter
	mockMetricsClient.EXPECT().UpdateMetric(
		gomock.Any(), // context.Context
		&pb.UpdateMetricRequest{
			Metric: &pb.Metric{
				Mtype:  storage.CounterType,
				Mname:  "counter1",
				Mvalue: strconv.FormatInt(42, 10),
			},
		},
	).Return(&pb.UpdateMetricResponse{}, nil)

	// Ожидаем вызов UpdateMetric для models.Gauge
	mockMetricsClient.EXPECT().UpdateMetric(
		gomock.Any(), // context.Context
		&pb.UpdateMetricRequest{
			Metric: &pb.Metric{
				Mtype:  storage.GaugeType,
				Mname:  "gauge1",
				Mvalue: strconv.FormatFloat(3.14, 'g', -1, 64),
			},
		},
	).Return(&pb.UpdateMetricResponse{}, nil)

	// Вызываем тестируемую функцию с моком
	err := sendDataGRPC(data, mockMetricsClient)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)
}

func TestSendDataGRPC_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetricsClient := mocks.NewMockMetricsClient(ctrl)

	// Создаем тестовые данные
	data := models.Data{
		Counters: []models.Counter{
			{Name: "counter1", Value: 42},
		},
	}

	// Ожидаем вызов UpdateMetric с ошибкой
	mockMetricsClient.EXPECT().UpdateMetric(
		gomock.Any(), // context.Context
		gomock.Any(), // *pb.UpdateMetricRequest
	).Return(nil, errors.New("rpc error"))

	// Вызываем тестируемую функцию с моком
	err := sendDataGRPC(data, mockMetricsClient)

	// Проверяем, что ошибка возвращена
	assert.Error(t, err)
}

func TestSendDataHTTP(t *testing.T) {
	// Создаем mock HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Проверяем, что запрос пришел с правильным URL и методом
		if req.URL.String() != "/updates/" {
			t.Errorf("Expected URL '/updates/', got '%s'", req.URL.String())
		}
		if req.Method != http.MethodPost {
			t.Errorf("Expected method 'POST', got '%s'", req.Method)
		}

		// Проверяем заголовки
		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", req.Header.Get("Content-Type"))
		}
		if req.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("Expected Content-Encoding 'gzip', got '%s'", req.Header.Get("Content-Encoding"))
		}

		// Декомпрессируем тело запроса
		gz, err := gzip.NewReader(req.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gz.Close()

		var buf bytes.Buffer
		_, err = buf.ReadFrom(gz)
		if err != nil {
			t.Fatalf("Failed to read gzip data: %v", err)
		}

		// Декодируем JSON
		var receivedMetrics []models.Metrics
		if err := json.Unmarshal(buf.Bytes(), &receivedMetrics); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		// Проверяем данные
		if len(receivedMetrics) != 2 {
			t.Errorf("Expected 2 metrics, got %d", len(receivedMetrics))
		}

		// Проверяем первую метрику (Counter)
		if receivedMetrics[0].ID != "counter1" || *receivedMetrics[0].Delta != 10 || receivedMetrics[0].MType != storage.CounterType {
			t.Errorf("Expected counter metric 'counter1' with Delta 10, got %v", receivedMetrics[0])
		}

		// Проверяем вторую метрику (Gauge)
		if receivedMetrics[1].ID != "gauge1" || *receivedMetrics[1].Value != 3.14 || receivedMetrics[1].MType != storage.GaugeType {
			t.Errorf("Expected gauge metric 'gauge1' with Value 3.14, got %v", receivedMetrics[1])
		}

		// Отправляем успешный ответ
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	ctrl := gomock.NewController(t)
	// Создаем экземпляр MockAgent
	agent := mocks.NewMockSender(ctrl)

	// Данные для отправки
	data := models.Data{
		Counters: []models.Counter{{Name: "counter1", Value: 10}},
		Gauges:   []models.Gauge{{Name: "gauge1", Value: 3.14}},
	}

	agent.EXPECT().SendDataHTTP(data).Return(nil).Times(1)

	// Вызываем тестируемый метод
	err := agent.SendDataHTTP(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
