package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Server represents the API server.
type Server struct {
	router *mux.Router
	addr   string
}

// NewServer creates a new API server.
func NewServer(addr string) *Server {
	server := &Server{
		router: mux.NewRouter(),
		addr:   addr,
	}
	server.setupRoutes()
	return server
}

// setupRoutes configures all API routes.
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/api/health", HealthHandler).Methods(http.MethodGet)

	// Downloads
	s.router.HandleFunc("/api/downloads", ListDownloadsHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/api/downloads", CreateDownloadHandler).Methods(http.MethodPost)
	s.router.HandleFunc("/api/downloads/{id}", GetDownloadHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/api/downloads/{id}", UpdateDownloadHandler).Methods(http.MethodPut)
	s.router.HandleFunc("/api/downloads/{id}", DeleteDownloadHandler).Methods(http.MethodDelete)
	s.router.HandleFunc("/api/downloads/status/{status}", ListDownloadsByStatusHandler).Methods(http.MethodGet)

	// Library
	s.router.HandleFunc("/api/library", ListLibraryHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/api/library/search", SearchLibraryHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/api/library/{id}", DeleteLibraryItemHandler).Methods(http.MethodDelete)

	// Settings
	s.router.HandleFunc("/api/settings", GetSettingsHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/api/settings", UpdateSettingsHandler).Methods(http.MethodPut)

	// WebSocket for real-time progress
	s.router.HandleFunc("/ws/progress", ProgressWebSocketHandler)
}

// ProgressWebSocketHandler handles WebSocket connections for progress updates.
func ProgressWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Keep connection open and ready to receive/send messages
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		// Process message and send updates back
	}
}

// Start starts the API server.
func (s *Server) Start() error {
	fmt.Printf("Starting API server on %s\n", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}

// GetRouter returns the mux router for testing.
func (s *Server) GetRouter() *mux.Router {
	return s.router
}
