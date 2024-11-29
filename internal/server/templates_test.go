package server

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	// Создаем экземпляр сервера и настраиваем рендерер
	s := &Server{router: echo.New()}
	s.ConfigureRenderer("../templates/metrics.html")

	// Создаем буфер для записи вывода
	var buf bytes.Buffer

	// Данные для шаблона
	data := map[string]interface{}{
		"Title":   "Test Title",
		"Message": "Test Message",
	}

	// Рендерим шаблон "index.html"
	err := s.router.Renderer.(*Template).Render(&buf, "metrics.html", data, nil)
	assert.NoError(t, err)

	// Проверяем, что вывод соответствует ожидаемому
	expected := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
</head>
<body>
    <h2>Counter Metrics</h2>
    <ul>
        
    </ul>
    <br>
    <h2>Gauge Metrics</h2>
    <ul>
        
    </ul>
</body>
</html>`
	assert.Equal(t, expected, buf.String())
}

func TestConfigureRenderer(t *testing.T) {
	// Создаем экземпляр сервера и настраиваем рендерер
	s := &Server{router: echo.New()}
	s.ConfigureRenderer("../templates/metrics.html")

	// Проверяем, что рендерер установлен
	assert.NotNil(t, s.router.Renderer)

	// Создаем тестовый HTTP-запрос
	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := s.router.NewContext(req, rec)

	// Данные для шаблона
	data := map[string]interface{}{
		"Title":   "Test Title",
		"Message": "Test Message",
	}

	// Рендерим шаблон "about.html"
	err := s.router.Renderer.(*Template).Render(rec, "metrics.html", data, c)
	assert.NoError(t, err)

	// Проверяем, что вывод соответствует ожидаемому
	expected := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
</head>
<body>
    <h2>Counter Metrics</h2>
    <ul>
        
    </ul>
    <br>
    <h2>Gauge Metrics</h2>
    <ul>
        
    </ul>
</body>
</html>`
	assert.Equal(t, expected, rec.Body.String())
}
