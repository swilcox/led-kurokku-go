package admin

import (
	"log/slog"
	"net/http"
	"time"
)

// Server is the kurokku-admin HTTP server.
type Server struct {
	store *Store
	mux   *http.ServeMux
}

// NewServer creates a new admin server with the given instance store.
func NewServer(store *Store) *Server {
	s := &Server{
		store: store,
		mux:   http.NewServeMux(),
	}
	initTemplates()
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.Handle("GET /static/", staticHandler())

	s.mux.HandleFunc("GET /{$}", s.handleIndex)
	s.mux.HandleFunc("GET /instances/new", s.handleInstanceNew)
	s.mux.HandleFunc("POST /instances", s.handleInstanceCreate)
	s.mux.HandleFunc("GET /instances/{id}/edit", s.handleInstanceEdit)
	s.mux.HandleFunc("PUT /instances/{id}", s.handleInstanceUpdate)
	s.mux.HandleFunc("DELETE /instances/{id}", s.handleInstanceDelete)
	s.mux.HandleFunc("POST /instances/{id}/test", s.handleInstanceTest)

	s.mux.HandleFunc("GET /instances/{id}/config", s.handleConfigView)
	s.mux.HandleFunc("GET /instances/{id}/config/preview", s.handlePreview)
	s.mux.HandleFunc("GET /instances/{id}/config/edit", s.handleConfigEdit)
	s.mux.HandleFunc("POST /instances/{id}/config", s.handleConfigSave)
	s.mux.HandleFunc("GET /instances/{id}/config/json", s.handleConfigJSON)
	s.mux.HandleFunc("POST /instances/{id}/config/json", s.handleConfigJSONSave)
	s.mux.HandleFunc("GET /instances/{id}/alert", s.handleAlertForm)
	s.mux.HandleFunc("POST /instances/{id}/alert", s.handleAlertSend)
	s.mux.HandleFunc("POST /instances/{id}/config/widgets/add", s.handleWidgetAdd)
	s.mux.HandleFunc("GET /config/widgets/form", s.handleWidgetForm)
	s.mux.HandleFunc("DELETE /instances/{id}/config/widgets/{idx}", s.handleWidgetRemove)
}

// responseRecorder wraps http.ResponseWriter to capture the status code.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

// ServeHTTP implements http.Handler with request logging.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
	s.mux.ServeHTTP(rr, r)
	slog.Info("http request",
		"method", r.Method,
		"path", r.URL.Path,
		"status", rr.status,
		"duration", time.Since(start),
	)
}
