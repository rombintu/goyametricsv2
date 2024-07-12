package server

import (
	"net/http"

	"github.com/rombintu/goyametricsv2/internal/storage"
)

type Server struct {
	storage *storage.Storage
	router  *http.ServeMux
}

func NewServer(storage *storage.Storage) *Server {
	return &Server{
		router:  http.NewServeMux(),
		storage: storage,
	}
}

func (s *Server) Start() {
	s.ConfigureRouter()
	s.ConfigureStorage()
	if err := http.ListenAndServe(":8080", s.router); err != nil {
		panic(err)
	}
}

func (s *Server) ConfigureStorage() {
	s.storage.Open()
}

func (s *Server) ConfigureRouter() {
	s.router.HandleFunc("/", s.MetricsHandler)
}
