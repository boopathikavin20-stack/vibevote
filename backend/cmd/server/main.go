package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"pulsevote/config"
	"pulsevote/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("mongo disconnect error: %v", err)
		}
	}()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}

	redisClient, err := config.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	router := routes.SetupRouter(cfg, client, redisClient)
	frontendDist := os.Getenv("FRONTEND_DIST")
	if frontendDist == "" {
		frontendDist = filepath.Join("..", "frontend", "dist")
	}
	frontendIndex := filepath.Join(frontendDist, "index.html")
	router.Static("/assets", filepath.Join(frontendDist, "assets"))
	router.StaticFile("/favicon.svg", filepath.Join(frontendDist, "favicon.svg"))
	router.StaticFile("/icons.svg", filepath.Join(frontendDist, "icons.svg"))
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(frontendIndex)
	})
	server := &http.Server{
		Addr:    cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("PulseVote server starting on %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
}
