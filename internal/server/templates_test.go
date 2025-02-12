package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	// Создаем временную директорию для шаблонов
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "internal", "templates")
	err := os.MkdirAll(templatesDir, 0755)
	require.NoError(t, err)

	// Создаем тестовый шаблон
	templateContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
</head>
<body>
    <h2>Counter Metrics</h2>
    <ul>
        {{range .Counters}}
        <li>{{.Name}}: {{.Value}}</li>
        {{end}}
    </ul>
    <br>
    <h2>Gauge Metrics</h2>
    <ul>
        {{range .Gauges}}
        <li>{{.Name}}: {{.Value}}</li>
        {{end}}
    </ul>
</body>
</html>`

	templatePath := filepath.Join(templatesDir, "metrics.html")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	// Инициализация сервера с правильным конфигом
	s := &Server{router: echo.New()}
	err = s.ConfigureRenderer(RendererConfig{
		TemplatesGlob: filepath.Join(templatesDir, "*.html"),
	})
	require.NoError(t, err)

	// Проверяем тип рендерера
	renderer, ok := s.router.Renderer.(*Template)
	require.True(t, ok, "Renderer should be of type *Template")
	require.NotNil(t, renderer.templates)

	t.Run("successful render", func(t *testing.T) {
		var buf bytes.Buffer
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := s.router.NewContext(req, rec)

		// Тестовые данные
		data := map[string]interface{}{
			"Title": "Metrics Report",
			"Counters": []struct {
				Name  string
				Value int
			}{
				{"Requests", 42},
				{"Errors", 3},
			},
			"Gauges": []struct {
				Name  string
				Value float64
			}{
				{"Memory", 65.3},
				{"CPU", 12.7},
			},
		}

		// Рендеринг шаблона
		err = renderer.Render(&buf, "metrics.html", data, c)
		require.NoError(t, err)

		// Проверка результата
		expected := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics Report</title>
</head>
<body>
    <h2>Counter Metrics</h2>
    <ul>
        <li>Requests: 42</li>
        <li>Errors: 3</li>
    </ul>
    <br>
    <h2>Gauge Metrics</h2>
    <ul>
        <li>Memory: 65.3</li>
        <li>CPU: 12.7</li>
    </ul>
</body>
</html>`
		assert.Equal(t, normalizeHTML(expected), normalizeHTML(buf.String()))
	})
}

// normalizeHTML удаляет пробелы и переносы для сравнения
func normalizeHTML(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func TestConfigureRenderer(t *testing.T) {
	// 1. Создаем тестовую директорию с шаблонами
	testTemplatesDir := filepath.Join("testdata", "templates")
	err := os.MkdirAll(testTemplatesDir, 0755)
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll("testdata") })

	// 2. Создаем тестовый шаблон
	testTemplate := filepath.Join(testTemplatesDir, "test.html")
	require.NoError(t, os.WriteFile(testTemplate, []byte(`
    <!DOCTYPE html>
    <html>
    <head><title>{{.Title}}</title></head>
    <body>{{.Message}}</body>
    </html>
    `), 0644))

	// 3. Инициализация сервера
	s := &Server{router: echo.New()}

	// 4. Конфигурация с правильным шаблоном
	err = s.ConfigureRenderer(RendererConfig{
		TemplatesGlob: filepath.Join(testTemplatesDir, "*.html"),
	})
	require.NoError(t, err)
	require.NotNil(t, s.router.Renderer)

	// 5. Безопасное приведение типа
	renderer, ok := s.router.Renderer.(*Template)
	require.True(t, ok, "Renderer should be of type *Template")
	require.NotNil(t, renderer.templates)

	// 6. Тестовый случай
	t.Run("successful render", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := s.router.NewContext(req, rec)

		data := map[string]interface{}{
			"Title":   "Test Page",
			"Message": "Hello World",
		}

		err = renderer.Render(rec, "test.html", data, c)
		require.NoError(t, err)

		assert.Contains(t, rec.Body.String(), "<title>Test Page</title>")
		assert.Contains(t, rec.Body.String(), "Hello World")
	})
}
