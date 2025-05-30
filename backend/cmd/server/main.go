package main

import (
	"fmt"
	"log"
	"net/http"
	"sleek-chat-backend/internal/adapters/handlers"
	"sleek-chat-backend/internal/adapters/middleware"
	"sleek-chat-backend/internal/domain/repository"
	"sleek-chat-backend/internal/domain/usecase"
	"sleek-chat-backend/internal/infrastructure/database"
	"sleek-chat-backend/internal/infrastructure/websocket"
	"sleek-chat-backend/pkg/config"
	"sleek-chat-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
    "github.com/swaggo/files"
    "sleek-chat-backend/cmd/server/docs"
)

// @title SleekChat API
// @version 1.0
// @description SleekChat backend with authentication and messaging
// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()

	appLogger := logger.New()
	appLogger.Info("Starting Sleek Chat Backend Server...")

	db, err := database.New(&cfg.Database)
	if err != nil {
		appLogger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		appLogger.Fatalf("Failed to migrate database: %v", err)
	}
	appLogger.Info("Database migration completed")
	repos := &repository.Repository{
		User:        database.NewUserRepository(db.DB),
		Chat:        database.NewChatRepository(db.DB),
		Message:     database.NewMessageRepository(db.DB),
		Session:     database.NewSessionRepository(db.DB),
		KeyExchange: database.NewKeyExchangeRepository(db.DB),
	}
	authUseCase := usecase.NewAuthUseCase(repos.User, repos.Session, cfg.JWT.Secret)
	userUseCase := usecase.NewUserUseCase(repos.User)
	keyExchangeUseCase := usecase.NewKeyExchangeUseCase(repos.Session, repos.User, appLogger)

	wsHub := websocket.NewHub(appLogger, nil)
	go wsHub.Run()

	chatUseCase := usecase.NewChatUseCase(repos.Chat, repos.Message, repos.User, repos.KeyExchange, wsHub)

	wsHub.SetChatUseCase(chatUseCase)

	authHandler := handlers.NewAuthHandler(authUseCase, appLogger)
	chatHandler := handlers.NewChatHandler(chatUseCase, wsHub, appLogger)
	userHandler := handlers.NewUserHandler(userUseCase, appLogger)
	wsHandler := handlers.NewWebSocketHandler(wsHub, appLogger)

	authMiddleware := middleware.NewAuthMiddleware(authUseCase, appLogger)
	encryptionMiddleware := middleware.NewEncryptionMiddleware(repos.Session, appLogger)
	keyExchangeHandler := handlers.NewKeyExchangeHandler(keyExchangeUseCase, encryptionMiddleware, appLogger)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware(appLogger))
	// Добавляем middleware для шифрования (применяется ко всем маршрутам)
	router.Use(encryptionMiddleware.DecryptRequest())
	router.Use(encryptionMiddleware.EncryptResponse())

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "sleek-chat-backend",
		})
	})

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
			auth.GET("/profile", authMiddleware.RequireAuth(), authHandler.GetProfile)
			auth.POST("/change-password", authMiddleware.RequireAuth(), authHandler.ChangePassword)
		}

		chats := api.Group("/chats")
		chats.Use(authMiddleware.RequireAuth())
		{
			chats.POST("", chatHandler.CreateChat)
			chats.POST("/private", chatHandler.CreateOrGetPrivateChat)
			chats.GET("", chatHandler.GetUserChats)
			chats.GET("/:id/messages", chatHandler.GetChatMessages)
			chats.POST("/:id/messages", chatHandler.SendMessage)
			chats.GET("/:id/members", chatHandler.GetChatMembers)
			chats.POST("/:id/members", chatHandler.AddMember)
			chats.DELETE("/:id/members/:userId", chatHandler.RemoveMember)
			chats.PUT("/:id/members/:userId/admin", chatHandler.SetAdmin)
			chats.DELETE("/:id/members/:userId/admin", chatHandler.RemoveAdmin)
			chats.POST("/:id/leave", chatHandler.LeaveChat)
			chats.DELETE("/:id", chatHandler.DeleteChat)
			chats.DELETE("/:id/delete", chatHandler.DeleteGroupChat)
		}
		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			users.GET("/search", userHandler.SearchUsers)
			users.GET("/online", userHandler.GetOnlineUsers)
			users.GET("/:id", userHandler.GetUser)
		}

		// Регистрируем маршруты для обмена ключами
		keyExchangeHandler.RegisterRoutesWithMiddleware(api, authMiddleware)

		api.GET("/ws", authMiddleware.WebSocketAuth(), wsHandler.HandleWebSocket)
	}

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
