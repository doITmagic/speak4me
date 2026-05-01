package api

import (
	"database/sql"
	"net/http"

	"github.com/doITmagic/speak4me/pkg/tts"
)

// APIServer grupează dependențele necesare pentru a deservi rutele HTTP
type APIServer struct {
	Registry *tts.ModelRegistry
	DB       *sql.DB
	Router   *http.ServeMux
}

// NewServer instanțiază serverul HTTP și rutează endpoint-urile
func NewServer(registry *tts.ModelRegistry, db *sql.DB) *APIServer {
	server := &APIServer{
		Registry: registry,
		DB:       db,
		Router:   http.NewServeMux(),
	}
	server.routes()
	return server
}

// routes atașează logica la path-urile specifice
func (s *APIServer) routes() {
	s.Router.HandleFunc("GET /api/v1/voices", s.handleGetVoices())
	s.Router.HandleFunc("POST /api/v1/synthesize", s.handleSynthesize())
}
