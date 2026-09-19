package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"pulsevote/config"
	"pulsevote/controllers"
	"pulsevote/middleware"
	"pulsevote/redis"
	"pulsevote/repository"
	"pulsevote/services"
	"pulsevote/websocket"
)

func SetupRouter(cfg config.Config, client *mongo.Client, redisClient *goredis.Client) *gin.Engine {
	mongoDB := client.Database("pulsevote")
	userRepo := repository.NewUserRepository(mongoDB)
	pollRepo := repository.NewPollRepository(mongoDB)
	voteRepo := repository.NewVoteRepository(mongoDB)

	if err := pollRepo.EnsureIndexes(context.Background()); err != nil {
		fmt.Println("poll index error:", err)
	}
	if err := voteRepo.EnsureIndexes(context.Background()); err != nil {
		fmt.Println("vote index error:", err)
	}

	store := redis.NewStore(redisClient)
	pollService := services.NewPollService(pollRepo, voteRepo, store)
	authService := services.NewAuthService(userRepo)
	authController := controllers.NewAuthController(authService)
	pollController := controllers.NewPollController(pollService)
	voteController := controllers.NewVoteController(pollService)
	wsHub := websocket.NewHub()

	middleware.SetJWTSecret(cfg.JWTSecret)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "PulseVote backend healthy", "data": gin.H{"status": "ok"}})
	})

	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/signup", authController.Signup)
		authRoutes.POST("/login", authController.Login)
		authRoutes.GET("/me", middleware.AuthRequired(), authController.Me)
	}

	pollRoutes := r.Group("/api/polls")
	{
		pollRoutes.POST("", middleware.AuthRequired(), pollController.Create)
		pollRoutes.GET("", middleware.AuthRequired(), pollController.List)
		pollRoutes.GET("/public/:shareCode", pollController.GetPublic)
		pollRoutes.GET("/:id/stats", middleware.AuthRequired(), pollController.Stats)
		pollRoutes.PUT("/:id", middleware.AuthRequired(), pollController.Update)
		pollRoutes.DELETE("/:id", middleware.AuthRequired(), pollController.Delete)
		pollRoutes.POST("/:id/close", middleware.AuthRequired(), pollController.Close)
		pollRoutes.POST("/:id/vote", voteController.Vote)
		pollRoutes.GET("/:id/ws", func(c *gin.Context) {
			pollID := c.Param("id")
			websocket.ServePollWebSocket(c, wsHub, pollID)
		})
	}

	go func() {
		pubsub := redisClient.Subscribe(context.Background(), "poll:updates")
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage(context.Background())
			if err != nil {
				continue
			}
			if msg != nil {
				wsHub.BroadcastPollUpdate(msg.Channel, []byte(msg.Payload))
			}
		}
	}()

	return r
}
