package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/collabb/backend/internal/auth"
	"github.com/collabb/backend/internal/board"
	"github.com/collabb/backend/internal/db"
	"github.com/collabb/backend/internal/middleware"
	"github.com/collabb/backend/internal/ws"
	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	hub := ws.NewHub()
	go hub.Run()

	authSvc := auth.NewService(pool)
	boardSvc := board.NewService(pool)

	r := mux.NewRouter()

	// Public routes
	auth.NewHandler(authSvc).RegisterRoutes(r)

	// Protected routes
	api := r.PathPrefix("").Subrouter()
	api.Use(middleware.Auth)
	board.NewHandler(boardSvc, hub).RegisterRoutes(api)

	// WebSocket endpoint
	api.HandleFunc("/ws/{boardID}", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWS(w, r, mux.Vars(r)["boardID"])
	}).Methods(http.MethodGet)

	addr := ":" + envOr("PORT", "8080")
	srv := &http.Server{
		Addr: addr,
		// Wrap the entire router with CORS so preflight OPTIONS requests
		// are handled before gorilla/mux routing (which would 404 them).
		Handler:      middleware.CORS(r),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down…")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	hub.Stop()
	log.Println("done")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
