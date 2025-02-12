package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/mocks"
	models "github.com/rombintu/goyametricsv2/internal/models"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/rombintu/goyametricsv2/lib/ptrhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	counterMetricType = "counter"
	gaugeMetricType   = "gauge"
)

func TestServer_updateMetrics(t *testing.T) {
	e := echo.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	s := NewServer(m, config.ServerConfig{})
	s.ConfigureRouter()

	m.EXPECT().Update(counterMetricType, "counter1", "1").Return(nil)
	m.EXPECT().Update(counterMetricType, "counter1", "5").Return(nil)
	m.EXPECT().Update(gaugeMetricType, "gauge1", "1.5").Return(nil)
	m.EXPECT().Update(gaugeMetricType, "gauge1", "2").Return(nil)

	type want struct {
		code        int
		response    string
		contentType string
	}
	type params struct {
		mtype  string
		mname  string
		mvalue string
	}
	tests := []struct {
		name   string
		want   want
		target params
	}{
		{
			name: "addNewCounter",
			want: want{
				code:        http.StatusOK,
				response:    "updated",
				contentType: echo.MIMETextHTML,
			},
			target: params{
				mtype:  counterMetricType,
				mname:  "counter1",
				mvalue: "1",
			},
		},
		{
			name: "addOldCounter",
			want: want{
				code:        http.StatusOK,
				response:    "updated",
				contentType: echo.MIMETextHTML,
			},
			target: params{
				mtype:  counterMetricType,
				mname:  "counter1",
				mvalue: "5",
			},
		},
		{
			name: "addNewGauge",
			want: want{
				code:        http.StatusOK,
				response:    "updated",
				contentType: echo.MIMETextHTML,
			},
			target: params{

				mtype:  gaugeMetricType,
				mname:  "gauge1",
				mvalue: "1.5",
			},
		},
		{
			name: "addOldGauge",
			want: want{
				code:        http.StatusOK,
				response:    "updated",
				contentType: echo.MIMETextHTML,
			},
			target: params{

				mtype:  gaugeMetricType,
				mname:  "gauge1",
				mvalue: "2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", echo.MIMETextHTML)
			c := e.NewContext(req, rec)
			c.SetPath("/update/:mtype/:mname/:mvalue")
			c.SetParamNames("mtype", "mname", "mvalue")
			c.SetParamValues(tt.target.mtype, tt.target.mname, tt.target.mvalue)

			// Check
			if assert.NoError(t, s.MetricsHandler(c)) {
				assert.Equal(t, tt.want.code, rec.Code)
				assert.Equal(t, tt.want.response, rec.Body.String())
				assert.Equal(t, tt.want.contentType, rec.Header().Get("Content-Type"))
			}
		})
	}
}

func TestServer_MetricGetHandler(t *testing.T) {
	e := echo.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	s := NewServer(m, config.ServerConfig{})
	s.ConfigureRouter()

	m.EXPECT().Get(counterMetricType, "counter1").Return("1", nil).AnyTimes()
	m.EXPECT().Get(counterMetricType, "unknown").Return("", errors.New("not found"))

	type want struct {
		code        int
		response    string
		contentType string
	}
	type params struct {
		mtype string
		mname string
	}
	tests := []struct {
		name   string
		want   want
		target params
	}{
		{
			name: "getKnownMetric",
			want: want{
				code:        http.StatusOK,
				response:    "1",
				contentType: echo.MIMETextHTML,
			},
			target: params{
				mtype: counterMetricType,
				mname: "counter1",
			},
		},
		{
			name: "getUnknownMetric",
			want: want{
				code:        http.StatusNotFound,
				response:    "not found",
				contentType: echo.MIMETextHTML,
			},
			target: params{
				mtype: counterMetricType,
				mname: "unknown",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			rec.Header().Set("Content-Type", echo.MIMETextHTML)
			c := e.NewContext(req, rec)
			c.SetPath("/value/:mtype/:mname")
			c.SetParamNames("mtype", "mname")
			c.SetParamValues(tt.target.mtype, tt.target.mname)

			if assert.NoError(t, s.MetricGetHandler(c)) {
				assert.Equal(t, tt.want.code, rec.Code)
				assert.Equal(t, tt.want.response, rec.Body.String())
				assert.Equal(t, tt.want.contentType, rec.Header().Get("Content-Type"))
			}
		})
	}
}

func TestServer_MetricUpdateHandlerJSON(t *testing.T) {
	e := echo.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	s := NewServer(m, config.ServerConfig{SyncMode: false})
	defer s.Shutdown()
	s.ConfigureRouter()

	validMetric := models.Metrics{
		ID:    "test_metric",
		MType: counterMetricType,
		Delta: ptrhelper.Int64Ptr(10),
	}

	invalidMetric := models.Metrics{
		ID:    "invalid_metric",
		MType: "invalid",
		Delta: ptrhelper.Int64Ptr(10),
	}

	t.Run("ValidMetric", func(t *testing.T) {
		m.EXPECT().Update(counterMetricType, validMetric.ID, "10").Return(nil)
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		body, _ := json.Marshal(validMetric)
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := s.MetricUpdateHandlerJSON(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Тест на невалидные данные
	t.Run("InvalidMetric", func(t *testing.T) {
		m.EXPECT().Update(invalidMetric.MType, invalidMetric.ID, "").Return(errors.New("invalid"))
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		body, _ := json.Marshal(invalidMetric)
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := s.MetricUpdateHandlerJSON(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestServer_MetricUpdatesHandlerJSON(t *testing.T) {
	e := echo.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	s := NewServer(m, config.ServerConfig{SyncMode: false})
	defer s.Shutdown()
	s.ConfigureRouter()

	counters := make(storage.Counters)
	counters["c1"] = 10

	gauges := make(storage.Gauges)
	gauges["g1"] = 10

	data := storage.Data{
		Counters: counters,
		Gauges:   gauges,
	}

	payload := []models.Metrics{
		{
			ID:    "c1",
			MType: counterMetricType,
			Delta: ptrhelper.Int64Ptr(10),
		},
		{
			ID:    "g1",
			MType: gaugeMetricType,
			Value: ptrhelper.Float64Ptr(10),
		},
	}

	t.Run("UpdateMetricsJSON", func(t *testing.T) {
		m.EXPECT().UpdateAll(data).Return(nil)
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := s.MetricUpdatesHandlerJSON(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

}

func TestServer_MetricValueHandlerJSON(t *testing.T) {
	e := echo.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	s := NewServer(m, config.ServerConfig{SyncMode: false})
	defer s.Shutdown()
	s.ConfigureRouter()

	payload := models.Metrics{
		ID:    "c1",
		MType: counterMetricType,
	}

	t.Run("GetMetricJSON", func(t *testing.T) {
		m.EXPECT().Get(counterMetricType, payload.ID).Return("10", nil)
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := s.MetricValueHandlerJSON(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

}

func TestServer_PingDatabase(t *testing.T) {
	e := echo.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorage(ctrl)
	s := NewServer(m, config.ServerConfig{SyncMode: false})
	defer s.Shutdown()
	s.ConfigureRouter()

	t.Run("PingDatabase", func(t *testing.T) {
		m.EXPECT().Ping().Return(nil).AnyTimes()
		m.EXPECT().Save().Return(nil).AnyTimes()
		m.EXPECT().Close().Return(nil).AnyTimes()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		// req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		err := s.PingDatabase(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestServer_RootHandler_HTMLContent(t *testing.T) {
	// Создаем временную директорию для шаблонов
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "internal", "templates")
	err := os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Создаем тестовый шаблон
	templateContent := `<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <div id="counters">
        {{range $name, $value := .Counters}}
        <p class="counter">{{$name}}: {{$value}}</p>
        {{end}}
    </div>
    <div id="gauges">
        {{range $name, $value := .Gauges}}
        <p class="gauge">{{$name}}: {{$value}}</p>
        {{end}}
    </div>
</body>
</html>`

	templatePath := filepath.Join(templatesDir, "metrics.html")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	// Инициализация Echo и моков
	e := echo.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	expectedData := storage.Data{
		Counters: storage.Counters{"c1": 1, "c2": 5},
		Gauges:   storage.Gauges{"g1": 2.3, "g2": 4.56},
	}

	mockStorage.EXPECT().GetAll().Return(expectedData).Times(1)

	// Инициализация сервера
	server := &Server{
		storage: mockStorage,
		router:  e,
	}

	// Настройка рендерера с временными шаблонами
	err = server.ConfigureRenderer(RendererConfig{
		TemplatesGlob: filepath.Join(templatesDir, "*.html"),
	})
	require.NoError(t, err)
	require.NotNil(t, server.router.Renderer)

	// Подготовка контекста
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Выполнение обработчика
	err = server.RootHandler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Получаем HTML-ответ
	htmlResponse := rec.Body.String()

	// Проверка счетчиков
	counterRegex := regexp.MustCompile(`<p class="counter">([^<]+): ([^<]+)</p>`)
	counterMatches := counterRegex.FindAllStringSubmatch(htmlResponse, -1)
	require.Equal(t, 2, len(counterMatches), "Expected 2 counters")

	expectedCounters := map[string]string{
		"c1": "1",
		"c2": "5",
	}
	for _, match := range counterMatches {
		name := match[1]
		value := match[2]
		expectedVal, ok := expectedCounters[name]
		if assert.True(t, ok, "Unexpected counter: %s", name) {
			assert.Equal(t, expectedVal, value)
		}
		delete(expectedCounters, name)
	}
	assert.Empty(t, expectedCounters, "Not all counters rendered")

	// Проверка измерений
	gaugeRegex := regexp.MustCompile(`<p class="gauge">([^<]+): ([^<]+)</p>`)
	gaugeMatches := gaugeRegex.FindAllStringSubmatch(htmlResponse, -1)
	require.Equal(t, 2, len(gaugeMatches), "Expected 2 gauges")

	expectedGauges := map[string]string{
		"g1": "2.3",
		"g2": "4.56",
	}
	for _, match := range gaugeMatches {
		name := match[1]
		value := match[2]
		expectedVal, ok := expectedGauges[name]
		if assert.True(t, ok, "Unexpected gauge: %s", name) {
			assert.Equal(t, expectedVal, value)
		}
		delete(expectedGauges, name)
	}
	assert.Empty(t, expectedGauges, "Not all gauges rendered")
}
