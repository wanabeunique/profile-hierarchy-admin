package routes

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"

	"test-app/backend/handlers"
	"test-app/backend/middleware"
)

func Setup(db *gorm.DB, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Handlers
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	profileHandler := handlers.NewProfileHandler(db)
	orderHandler := handlers.NewOrderHandler(db)

	// Public routes
	r.Post("/api/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtSecret))

		// Profiles
		r.Get("/api/profiles", profileHandler.List)
		r.Post("/api/profiles", profileHandler.Create)
		r.Get("/api/profiles/{id}", profileHandler.Get)
		r.Put("/api/profiles/{id}", profileHandler.Update)
		r.Delete("/api/profiles/{id}", profileHandler.Delete)

		// Orders (nested under profiles)
		r.Get("/api/profiles/{id}/orders", orderHandler.ListByProfile)
		r.Post("/api/profiles/{id}/orders", orderHandler.Create)

		// Orders (direct)
		r.Put("/api/orders/{id}", orderHandler.Update)
		r.Delete("/api/orders/{id}", orderHandler.Delete)
		r.Patch("/api/orders/{id}/pay", orderHandler.Pay)
	})

	return r
}
