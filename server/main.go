package main

import (
	"log"
	"net/http"

	"helpdesk/server/config"
	"helpdesk/server/db"
	"helpdesk/server/handlers"
	"helpdesk/server/middleware"
	"helpdesk/server/models"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("connect to DB: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Printf("migrations warning (may already exist): %v", err)
	}

	// Stores
	userStore := models.NewUserStore(database)
	ticketStore := models.NewTicketStore(database)
	commentStore := models.NewCommentStore(database)

	// Handlers
	authH := &handlers.AuthHandler{Users: userStore, JWTSecret: cfg.JWTSecret}
	userH := &handlers.UserHandler{Users: userStore}
	ticketH := &handlers.TicketHandler{Tickets: ticketStore}
	commentH := &handlers.CommentHandler{Comments: commentStore, Tickets: ticketStore}

	// Middleware chains
	authMW := middleware.Auth(cfg.JWTSecret)
	adminOnly := middleware.RequireRole(models.RoleAdmin)
	staffOnly := middleware.RequireRole(models.RoleAdmin, models.RoleOperator)

	mux := http.NewServeMux()

	// Auth (public)
	mux.HandleFunc("POST /api/auth/register", authH.Register)
	mux.HandleFunc("POST /api/auth/login", authH.Login)

	// Users — admin only
	mux.Handle("GET /api/users",
		authMW(adminOnly(http.HandlerFunc(userH.List))))
	mux.Handle("PUT /api/users/{id}/role",
		authMW(adminOnly(http.HandlerFunc(userH.UpdateRole))))
	mux.Handle("DELETE /api/users/{id}",
		authMW(adminOnly(http.HandlerFunc(userH.Delete))))

	// Tickets
	mux.Handle("POST /api/tickets",
		authMW(http.HandlerFunc(ticketH.Create)))
	mux.Handle("GET /api/tickets",
		authMW(http.HandlerFunc(ticketH.List)))
	mux.Handle("GET /api/tickets/{id}",
		authMW(http.HandlerFunc(ticketH.Get)))
	mux.Handle("PUT /api/tickets/{id}",
		authMW(http.HandlerFunc(ticketH.Update)))
	mux.Handle("DELETE /api/tickets/{id}",
		authMW(adminOnly(http.HandlerFunc(ticketH.Delete))))

	// Comments
	mux.Handle("POST /api/tickets/{id}/comments",
		authMW(http.HandlerFunc(commentH.Create)))
	mux.Handle("GET /api/tickets/{id}/comments",
		authMW(http.HandlerFunc(commentH.List)))
	mux.Handle("DELETE /api/comments/{id}",
		authMW(staffOnly(http.HandlerFunc(commentH.Delete))))

	handler := corsMiddleware(mux)

	log.Printf("server listening on :%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
