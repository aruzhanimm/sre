package main

import (
	"context"
	"net/http"
	"time"

	"BetKZ/internal/handlers"
	"BetKZ/internal/metrics"
	"BetKZ/internal/middleware"
	"BetKZ/internal/repository"
	"BetKZ/internal/service"
	ws "BetKZ/internal/websocket"
	jwtpkg "BetKZ/pkg/jwt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SetupRouter creates and configures the Gin router with all routes wired up.
// Separated from main() so it can be reused in E2E tests.
func SetupRouter(db *pgxpool.Pool, rdb *redis.Client, jwtSecret string, corsOrigins string) *gin.Engine {
	jwtManager := jwtpkg.NewJWTManager(jwtSecret, 15*time.Minute, 168*time.Hour)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	sportRepo := repository.NewSportRepository(db)
	eventRepo := repository.NewEventRepository(db)
	marketRepo := repository.NewMarketRepository(db)
	oddsRepo := repository.NewOddsRepository(db)
	betRepo := repository.NewBetRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	// WebSocket
	hub := ws.NewHub()
	go hub.Run()

	// Services
	authService := service.NewAuthService(userRepo, jwtManager)
	eventService := service.NewEventService(eventRepo, sportRepo)
	oddsService := service.NewOddsService(oddsRepo, marketRepo)
	betService := service.NewBetService(betRepo, userRepo, oddsRepo, marketRepo, txRepo, oddsService, hub)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	eventHandler := handlers.NewEventHandler(eventService)
	marketHandler := handlers.NewMarketHandler(oddsService)
	betHandler := handlers.NewBetHandler(betService)
	wsHandler := handlers.NewWSHandler(hub)

	// Router
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})
	r.GET("/health/db", func(c *gin.Context) {
		if err := db.Ping(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/health/redis", func(c *gin.Context) {
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Monitoring
	promMetrics, promRegistry := metrics.NewPrometheusMetrics(db)
	r.Use(promMetrics.GinMiddleware())
	r.GET("/metrics", gin.WrapH(metrics.MetricsHandler(promRegistry)))

	// WebSocket
	r.GET("/ws", wsHandler.HandleWebSocket)

	// Public API
	api := r.Group("/api")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.RefreshToken)
		api.GET("/sports", eventHandler.ListSports)
		api.GET("/events", eventHandler.ListEvents)
		api.GET("/events/:id", eventHandler.GetEvent)
		api.GET("/events/:id/markets", marketHandler.GetEventMarkets)
		api.GET("/markets/:id/odds", marketHandler.GetMarketOdds)
	}

	// Authenticated API
	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware(jwtManager))
	{
		auth.GET("/auth/profile", authHandler.GetProfile)
		auth.POST("/auth/logout", authHandler.Logout)
		auth.POST("/bets", betHandler.PlaceBet)
		auth.GET("/bets", betHandler.ListBets)
		auth.GET("/bets/:id", betHandler.GetBet)
		auth.GET("/transactions", betHandler.ListTransactions)
	}

	// Admin API
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.POST("/events", eventHandler.CreateEvent)
		admin.PUT("/events/:id", eventHandler.UpdateEvent)
		admin.PATCH("/events/:id/status", eventHandler.UpdateEventStatus)
		admin.DELETE("/events/:id", eventHandler.DeleteEvent)
		admin.POST("/markets", marketHandler.CreateMarket)
		admin.PUT("/odds/:id", marketHandler.ManualOverride)
		admin.GET("/odds/:id/history", marketHandler.GetOddsHistory)
		admin.GET("/bets", betHandler.AdminListBets)
		admin.POST("/settle", betHandler.SettleMarket)
		admin.GET("/stats", betHandler.GetDashboardStats)
		admin.POST("/users/:id/deposit", betHandler.Deposit)
		admin.POST("/deposit-by-email", betHandler.DepositByEmail)
		admin.POST("/place-bet", betHandler.AdminPlaceBet)
	}

	return r
}
