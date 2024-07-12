package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rombintu/goyametricsv2/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	counterMetricType = "counter"
	gaugeMetricType   = "gauge"
)

func TestServer_updateMetrics(t *testing.T) {
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
				response:    "",
				contentType: "text/plain; charset=utf-8",
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
				response:    "",
				contentType: "text/plain; charset=utf-8",
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
				response:    "",
				contentType: "text/plain; charset=utf-8",
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
				response:    "",
				contentType: "text/plain; charset=utf-8",
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
			url := fmt.Sprintf("/update/%s/%s/%s", tt.target.mtype, tt.target.mname, tt.target.mvalue)
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set("Content-Type", "text/plain; charset=utf-8")
			rec := httptest.NewRecorder()
			s.MetricsHandler(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
