// Package myorigin предоставляет middleware для проверки IP-адресов входящих запросов
// на принадлежность к доверенной подсети. Это может быть полезно для ограничения доступа
// к API или другим ресурсам только для определённых IP-адресов или диапазонов.
package myorigin

import (
	"errors"
	"net"

	"github.com/labstack/echo/v4"
)

// OriginMiddleware создает middleware для Echo, которое проверяет, принадлежит ли IP-адрес
// входящего запроса к указанной доверенной подсети. Если IP-адрес не принадлежит подсети,
// запрос завершается с ошибкой 403 Forbidden.
//
// Параметр trustedSubnet должен быть строкой в формате CIDR (например, "192.168.1.0/24").
// Если IP-адрес не может быть распознан или подсеть указана неверно, middleware возвращает
// соответствующую ошибку.
//
// Пример использования:
//
//	e := echo.New()
//	e.Use(OriginMiddleware("192.168.1.0/24"))
func OriginMiddleware(trustedSubnet string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Получаем IP-адрес из заголовка X-Real-IP
			ipStr := c.Request().Header.Get(echo.HeaderXRealIP)
			ipParsed := net.ParseIP(ipStr)
			if ipParsed == nil {
				c.Error(errors.New("invalid ip address"))
			}

			// Парсим доверенную подсеть
			_, ipv4Net, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				c.Error(err)
			}

			// Проверяем, принадлежит ли IP-адрес доверенной подсети
			if !ipv4Net.Contains(ipParsed) {
				c.String(echo.ErrForbidden.Code, echo.ErrForbidden.Error())
				return nil
			}

			// Продолжаем выполнение следующего обработчика
			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}
