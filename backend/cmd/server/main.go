package main

import (
	"crypto-chat-backend/internal/adapters/handlers"
	"crypto-chat-backend/internal/adapters/middleware"
	"crypto-chat-backend/internal/domain/repository"
	"crypto-chat-backend/internal/domain/usecase"
	"crypto-chat-backend/internal/infrastructure/database"
	"crypto-chat-backend/internal/infrastructure/websocket"
	"crypto-chat-backend/pkg/config"
	"crypto-chat-backend/pkg/logger"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем логгер
	appLogger := logger.New()
	appLogger.Info("Starting Crypto Chat Backend Server...")

	// Подключаемся к базе данных
	db, err := database.New(&cfg.Database)
	if err != nil {
		appLogger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Выполняем миграции
	if err := db.Migrate(); err != nil {
		appLogger.Fatalf("Failed to migrate database: %v", err)
	}
	appLogger.Info("Database migration completed")

	// Инициализируем репозитории
	repos := &repository.Repository{
		User:        database.NewUserRepository(db.DB),
		Chat:        database.NewChatRepository(db.DB),
		Message:     database.NewMessageRepository(db.DB),
		Session:     database.NewSessionRepository(db.DB),
		KeyExchange: database.NewKeyExchangeRepository(db.DB),
	}

	// Инициализируем use cases
	authUseCase := usecase.NewAuthUseCase(repos.User, repos.Session, cfg.JWT.Secret)
	chatUseCase := usecase.NewChatUseCase(repos.Chat, repos.Message, repos.User, repos.KeyExchange)

	// Инициализируем WebSocket hub
	wsHub := websocket.NewHub(appLogger)
	go wsHub.Run()

	// Инициализируем handlers
	authHandler := handlers.NewAuthHandler(authUseCase, appLogger)
	chatHandler := handlers.NewChatHandler(chatUseCase, wsHub, appLogger)
	wsHandler := handlers.NewWebSocketHandler(wsHub, appLogger)

	// Инициализируем middleware
	authMiddleware := middleware.NewAuthMiddleware(authUseCase, appLogger)

	// Настраиваем Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware(appLogger))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "crypto-chat-backend",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
			auth.GET("/profile", authMiddleware.RequireAuth(), authHandler.GetProfile)
		}

		// Chat routes
		chats := api.Group("/chats")
		chats.Use(authMiddleware.RequireAuth())
		{
			chats.POST("", chatHandler.CreateChat)
			chats.GET("", chatHandler.GetUserChats)
			chats.GET("/:id/messages", chatHandler.GetChatMessages)
			chats.POST("/:id/messages", chatHandler.SendMessage)
			chats.POST("/:id/members", chatHandler.AddMember)
			chats.DELETE("/:id/members/:userId", chatHandler.RemoveMember)
		}

		// WebSocket route
		api.GET("/ws", authMiddleware.RequireAuth(), wsHandler.HandleWebSocket)
	}

	// Запускаем сервер
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	appLogger.Infof("Server starting on %s", serverAddr)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
