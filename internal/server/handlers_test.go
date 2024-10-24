package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/mocks"
	"github.com/stretchr/testify/assert"
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

	m.EXPECT().Get(counterMetricType, "counter1").Return("1", nil)
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
