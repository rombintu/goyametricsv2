package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/stretchr/testify/assert"
)

const (
	counterMetricType = "counter"
	gaugeMetricType   = "gauge"
)

func TestServer_updateMetrics(t *testing.T) {
	e := echo.New()
	storage := storage.NewStorage("mem")
	s := NewServer(storage)
	s.ConfigureStorage()
	s.ConfigureRouter()
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
			name: "add new counter",
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
			name: "add old counter",
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
			name: "add new gauge",
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
			name: "add old gauge",
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
	storage := storage.NewStorage("mem")
	s := NewServer(storage)
	s.ConfigureStorage()
	s.ConfigureRouter()
	storage.Driver.Update(counterMetricType, "counter1", "1")
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
			name: "get known metric",
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
			name: "get unknown metric",
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
