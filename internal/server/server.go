package server

import (
	"net/http"
	"strings"

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
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if r.Method == http.MethodPost {
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			if strings.HasPrefix(path, "/update/") {
				parts := strings.Split(strings.TrimPrefix(path, "/update/"), "/")
				if len(parts) == 3 {
					metricType := parts[0]
					metricName := parts[1]
					metricValue := parts[2]
					if metricName == "" {
						http.Error(w, "Missing metric name", http.StatusNotFound)
						return
					}
					if err := s.storage.Driver.Update(metricType, metricName, metricValue); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					//http.StatusOK
					w.WriteHeader(http.StatusOK)
				} else {
					http.Error(w, "Invalid format", http.StatusNotFound)
				}

			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
		}
	})
}
