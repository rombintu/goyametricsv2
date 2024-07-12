package server

import (
	"io"
	"net/http"
	"text/template"

	"github.com/labstack/echo"
	"github.com/rombintu/goyametricsv2/internal/config"
	"github.com/rombintu/goyametricsv2/internal/storage"
)

type Server struct {
	config  config.ServerConfig
	storage *storage.Storage
	router  *echo.Echo
}

func NewServer(storage *storage.Storage, config config.ServerConfig) *Server {
	return &Server{
		config:  config,
		router:  echo.New(),
		storage: storage,
	}
}

func (s *Server) Start() {
	s.ConfigureRenderer()
	s.ConfigureRouter()
	s.ConfigureStorage()
	if err := http.ListenAndServe(":8080", s.router); err != nil {
		panic(err)
	}
}

func (s *Server) ConfigureStorage() {
	s.storage.Open()
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (s *Server) ConfigureRenderer() {
	t := &Template{
		templates: template.Must(template.ParseGlob("internal/templates/*.html")),
	}
	s.router.Renderer = t
}

func (s *Server) ConfigureRouter() {
	s.router.GET("/", s.RootHandler)
	s.router.GET("/value/:mtype/:mname", s.MetricGetHandler)
	s.router.POST("/update/:mtype/:mname/:mvalue", s.MetricsHandler)
}
