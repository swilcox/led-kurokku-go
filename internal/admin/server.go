package admin

import "net/http"

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
	s.mux.HandleFunc("GET /instances/{id}/config/edit", s.handleConfigEdit)
	s.mux.HandleFunc("POST /instances/{id}/config", s.handleConfigSave)
	s.mux.HandleFunc("GET /instances/{id}/config/json", s.handleConfigJSON)
	s.mux.HandleFunc("POST /instances/{id}/config/json", s.handleConfigJSONSave)
	s.mux.HandleFunc("POST /instances/{id}/config/widgets/add", s.handleWidgetAdd)
	s.mux.HandleFunc("DELETE /instances/{id}/config/widgets/{idx}", s.handleWidgetRemove)
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
