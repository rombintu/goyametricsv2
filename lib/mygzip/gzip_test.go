package mygzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func Test_gzipResponseWriter_Write(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		Writer         io.Writer
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "simple_write_2_RW",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				Writer:         &bytes.Buffer{},
			},
			args:    args{b: []byte("hello")},
			want:    5,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grw := &gzipResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				Writer:         tt.fields.Writer,
			}
			got, err := grw.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("gzipResponseWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("gzipResponseWriter.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGzipMiddleware(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Определяем тестовую функцию-обработчик
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}

	// Создаем тестовый запрос с поддержкой gzip
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAcceptEncoding, "gzip")

	// Создаем тестовый ResponseRecorder
	rec := httptest.NewRecorder()

	// Создаем контекст Echo
	c := e.NewContext(req, rec)

	// Применяем middleware
	middleware := GzipMiddleware(handler)

	// Вызываем обработчик с примененным middleware
	err := middleware(c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем, что ответ был сжат gzip
	if rec.Header().Get(echo.HeaderContentEncoding) != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got: %s", rec.Header().Get(echo.HeaderContentEncoding))
	}

	// Читаем сжатые данные из ответа
	gzipReader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	// Декомпрессируем данные
	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	// Проверяем, что декомпрессированные данные соответствуют ожидаемым
	expectedData := "Hello, World!"
	if string(decompressedData) != expectedData {
		t.Errorf("Expected response: %s, got: %s", expectedData, string(decompressedData))
	}
}

func TestGzipMiddlewareWithGzipRequest(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Определяем тестовую функцию-обработчик
	handler := func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(body))
	}

	// Создаем тестовый запрос с gzip-сжатием
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	gzipWriter.Write([]byte("Hello, World!"))
	gzipWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
	req.Header.Set(echo.HeaderContentEncoding, "gzip")

	// Создаем тестовый ResponseRecorder
	rec := httptest.NewRecorder()

	// Создаем контекст Echo
	c := e.NewContext(req, rec)

	// Применяем middleware
	middleware := GzipMiddleware(handler)

	// Вызываем обработчик с примененным middleware
	err := middleware(c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем, что ответ был сжат gzip
	if rec.Header().Get(echo.HeaderContentEncoding) != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got: %s", rec.Header().Get(echo.HeaderContentEncoding))
	}

	// Читаем сжатые данные из ответа
	gzipReader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	// Декомпрессируем данные
	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	// Проверяем, что декомпрессированные данные соответствуют ожидаемым
	expectedData := "Hello, World!"
	if string(decompressedData) != expectedData {
		t.Errorf("Expected response: %s, got: %s", expectedData, string(decompressedData))
	}
}
