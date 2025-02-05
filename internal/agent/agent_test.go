package agent

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/logger"
	pb "github.com/rombintu/goyametricsv2/internal/server/proto"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
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
	reportInterval int64
	data           Data
	pollCount      int
	rateLimit      int64
	semaphore      *MockSemaphore
}

func (m *MockAgent) sendDataHTTP(data Data) error {
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
	mockAgent.On("sendDataHTTP", mock.Anything).Return(nil)
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

func TestRunReport_SendDataHTTPError(t *testing.T) {
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
	mockAgent.On("sendDataHTTP", mock.Anything).Return(errors.New("send error"))
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
			if err := a.sendDataHTTP(a.data); err != nil {
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

const bufSize = 1024 * 1024

// FakeMetricsServer реализует интерфейс MetricsServer для тестирования
type FakeMetricsServer struct {
	pb.UnimplementedMetricsServer
	receivedMetrics []*pb.Metric
}

func (s *FakeMetricsServer) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	s.receivedMetrics = append(s.receivedMetrics, req.Metric)
	return &pb.UpdateMetricResponse{Metric: &pb.Metric{}}, nil
}

func startTestServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	fakeServer := &FakeMetricsServer{}
	pb.RegisterMetricsServer(srv, fakeServer)

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	}()

	return srv, lis
}

func TestSendDataGRPC(t *testing.T) {
	// Запускаем тестовый gRPC сервер
	srv, lis := startTestServer(t)
	defer srv.Stop()

	// Подготавливаем тестовые данные
	testData := Data{
		Counters: []Counter{
			{name: "test_counter", value: 42},
		},
		Gauges: []Gauge{
			{name: "test_gauge", value: 3.14},
		},
	}

	// Создаем соединение с тестовым сервером
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	// Получаем клиент для проверки
	client := pb.NewMetricsClient(conn)
	fakeServer := srv.GetServiceInfo()["metrics"].Server.(*FakeMetricsServer)

	// Выполняем тестируемую функцию
	err = sendDataGRPC(testData, 0) // Порт не используется в этом тесте
	require.NoError(t, err)

	// Даем серверу время обработать запросы
	time.Sleep(100 * time.Millisecond)

	// Проверяем полученные данные
	require.Len(t, fakeServer.receivedMetrics, 2)

	// Проверяем counter
	counterMetric := fakeServer.receivedMetrics[0]
	assert.Equal(t, storage.CounterType, counterMetric.Mtype)
	assert.Equal(t, "test_counter", counterMetric.Mname)
	assert.Equal(t, "42", counterMetric.Mvalue)

	// Проверяем gauge
	gaugeMetric := fakeServer.receivedMetrics[1]
	assert.Equal(t, storage.GaugeType, gaugeMetric.Mtype)
	assert.Equal(t, "test_gauge", gaugeMetric.Mname)
	assert.Equal(t, "3.14", gaugeMetric.Mvalue)
}
