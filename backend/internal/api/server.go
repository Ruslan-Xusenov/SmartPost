package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-telegram/bot"
	"github.com/smartpost/backend/internal/config"
	"github.com/smartpost/backend/internal/scheduler"
)

// Server is the HTTP API server for the TWA frontend.
type Server struct {
	router    *chi.Mux
	db        *pgxpool.Pool
	cfg       *config.Config
	publisher *scheduler.Publisher
	botAPI    *bot.Bot
}

// NewServer creates a new API server.
func NewServer(cfg *config.Config, db *pgxpool.Pool, publisher *scheduler.Publisher, webhookHandler http.Handler, botAPI *bot.Bot) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		db:        db,
		cfg:       cfg,
		publisher: publisher,
		botAPI:    botAPI,
	}
	s.setupMiddleware()
	s.setupRoutes(webhookHandler)
	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(chimw.Logger)
	s.router.Use(chimw.Recoverer)
	s.router.Use(chimw.RealIP)
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.cfg.TWAURL, "https://securehub.uz", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Telegram-Init-Data"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes(webhookHandler http.Handler) {
	if webhookHandler != nil {
		s.router.Post("/bot/webhook", webhookHandler.ServeHTTP)
	}

	s.router.Route("/api", func(r chi.Router) {
		// TWA auth middleware
		r.Use(s.twaAuthMiddleware)

		r.Get("/channels", s.listChannels)
		r.Post("/posts", s.createPost)
		r.Get("/posts", s.listPosts)
		r.Get("/posts/{id}", s.getPost)
		r.Put("/posts/{id}", s.updatePost)
		r.Delete("/posts/{id}", s.deletePost)
		r.Post("/posts/{id}/send", s.sendPost)
		r.Post("/posts/{id}/schedule", s.schedulePost)
		
		r.Post("/upload", s.uploadMedia)
	})

	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}

// Router returns the chi.Mux for mounting on an HTTP server.
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Start starts the API server.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.APIPort)
	log.Printf("🌐 API server starting on %s", addr)
	return http.ListenAndServe(addr, s.router)
}
