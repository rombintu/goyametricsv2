package agent

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	"github.com/rombintu/goyametricsv2/internal/mocks"
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

func Test_loadPSUtilsMetrics(t *testing.T) {
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
func loadPSUtilsMetricsDI(memProv MemoryProvider, cpuProv CPUProvider) Data {
	v, err := memProv.VirtualMemory()
	if err != nil {
		logger.Log.Warn(err.Error())
		return Data{}
	}

	u, err := cpuProv.Percent(0, false)
	if err != nil {
		logger.Log.Warn(err.Error())
		return Data{}
	}

	return Data{
		Gauges: []Gauge{
			{name: "TotalMemory", value: float64(v.Total)},
			{name: "FreeMemory", value: float64(v.Free)},
			{name: "CPUutilization1", value: u[0]},
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
	assert.Contains(t, result.Gauges, Gauge{name: "TotalMemory", value: 1000000000})
	assert.Contains(t, result.Gauges, Gauge{name: "FreeMemory", value: 500000000})
	assert.Contains(t, result.Gauges, Gauge{name: "CPUutilization1", value: 25.5})
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
	data := Data{
		Counters: []Counter{
			{name: "counter1", value: 42},
		},
		Gauges: []Gauge{
			{name: "gauge1", value: 3.14},
		},
	}

	// Ожидаем вызов UpdateMetric для Counter
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

	// Ожидаем вызов UpdateMetric для Gauge
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
	data := Data{
		Counters: []Counter{
			{name: "counter1", value: 42},
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
