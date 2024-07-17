package server

import (
	"net/http"

	"github.com/labstack/echo"
)

func (s *Server) MetricsHandler(c echo.Context) error {
	mtype := c.Param("mtype")
	mname := c.Param("mname")
	mvalue := c.Param("mvalue")

	if mname == "" {
		return c.String(http.StatusNotFound, "Missing metric name")
	}
	if err := s.storage.Driver.Update(mtype, mname, mvalue); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "updated")
}

func (s *Server) MetricGetHandler(c echo.Context) error {
	mtype := c.Param("mtype")
	mname := c.Param("mname")
	value, err := s.storage.Driver.Get(mtype, mname)
	if err != nil {
		return c.String(http.StatusNotFound, "not found")
	}
	return c.String(http.StatusOK, value)
}

func (s *Server) RootHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "metrics.html", s.storage.Driver.GetAll())
}
