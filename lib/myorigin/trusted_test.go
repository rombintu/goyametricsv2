package myorigin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestOriginMiddleware_TrustedSubnet(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Определяем доверенную подсеть
	trustedSubnet := "192.168.1.0/24"

	// Добавляем middleware в Echo
	e.Use(OriginMiddleware(trustedSubnet))

	// Создаем тестовый маршрут
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	// Создаем тестовый запрос с доверенным IP
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderXRealIP, "192.168.1.100")

	// Создаем ResponseRecorder для записи ответа
	rec := httptest.NewRecorder()

	// Выполняем запрос
	e.ServeHTTP(rec, req)

	// Проверяем, что ответ успешный
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOriginMiddleware_UntrustedSubnet(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Определяем доверенную подсеть
	trustedSubnet := "192.168.1.0/24"

	// Добавляем middleware в Echo
	e.Use(OriginMiddleware(trustedSubnet))

	// Создаем тестовый маршрут
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	// Создаем тестовый запрос с недоверенным IP
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderXRealIP, "10.0.0.1")

	// Создаем ResponseRecorder для записи ответа
	rec := httptest.NewRecorder()

	// Выполняем запрос
	e.ServeHTTP(rec, req)

	// Проверяем, что ответ запрещен
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, echo.ErrForbidden.Error(), rec.Body.String())
}

func TestOriginMiddleware_InvalidIP(t *testing.T) {
	// Создаем экземпляр Echo
	e := echo.New()

	// Определяем доверенную подсеть
	trustedSubnet := "192.168.1.0/24"

	// Добавляем middleware в Echo
	e.Use(OriginMiddleware(trustedSubnet))

	// Создаем тестовый маршрут
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	// Создаем тестовый запрос с невалидным IP
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderXRealIP, "invalid-ip")

	// Создаем ResponseRecorder для записи ответа
	rec := httptest.NewRecorder()

	// Выполняем запрос
	e.ServeHTTP(rec, req)

	// Проверяем, что ответ содержит ошибку
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
