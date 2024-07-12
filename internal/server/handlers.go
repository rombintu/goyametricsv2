package server

import (
	"net/http"
	"strings"
)

func (s *Server) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	if r.Method == http.MethodPost {
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
}
