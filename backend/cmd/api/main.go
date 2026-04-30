package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"BetKZ/config"
	"BetKZ/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	// Connect databases
	db := database.Connect(cfg.DatabaseURL())
	defer db.Close()
	rdb := database.ConnectRedis(cfg.RedisURL)
	defer rdb.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup router using shared function
	r := SetupRouter(db, rdb, cfg.JWTSecret, cfg.CORSOrigins)

	// Start server
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		fmt.Printf("🚀 BetKZ API server starting on :%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}
