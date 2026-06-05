package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"goapp/handlers"
	"goapp/middleware"
	"goapp/models"
	"goapp/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Swagger docs
	_ "goapp/docs"
)

// @title           Lab Project API
// @version         1.0
// @description     Документация API для лабораторных работ №2-№4
// @description     Реализована аутентификация через JWT (Cookies) и OAuth2 (Yandex)
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@labproject.com

// @license.name   MIT
// @license.url    https://opensource.org/licenses/MIT

// @host           localhost:8080
// @BasePath       /

// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
// @description JWT токен доступа хранится в HttpOnly cookie. Для авторизации в Swagger UI выполните /auth/login, после чего cookie будет отправляться автоматически.

// @securityDefinitions.oauth2.authorizationcode OAuth2Yandex
// @tokenUrl https://oauth.yandex.ru/token
// @authorizationUrl https://oauth.yandex.ru/authorize
// @scope.read Режим чтения
// @scope.write Режим записи
// @description OAuth2 аутентификация через Яндекс

func main() {
	// ================= ENV =================
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using environment variables")
	}

	// ================= DB =================
	// Настройка подключения к БД с повторными попытками
	var db *gorm.DB
	maxRetries := 5
	retryDelay := 3 * time.Second

	// Используем стандартный формат DSN
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	log.Println("Connecting to database...")

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatal("Failed to connect database after retries:", err)
	}

	log.Println("Connected to DB")

	// Получение sql.DB для проверки соединения
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}
	defer sqlDB.Close()

	// Проверка соединения
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// UUID extension
	err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`).Error
	if err != nil {
		log.Fatal("failed to enable pgcrypto:", err)
	}

	// ================= MIGRATIONS =================
	err = db.AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.RefreshToken{},
	)

	if err != nil {
		log.Fatal("failed to migrate:", err)
	}

	log.Println("Database migrated")

	// ================= SERVICES =================
	authService := services.NewAuthService(db)

	// ================= ROUTER =================
	r := gin.Default()

	// ================= CORS =================
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:4200",
			"http://localhost:8080",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		AllowCredentials: true,
	}))

	// ================= SWAGGER (только для development) =================
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "development" || appEnv == "local" || appEnv == "" {
		r.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
		log.Println("✅ Swagger UI доступен на http://localhost:8080/api/docs/index.html")
	} else {
		log.Println("⚠️ Swagger UI отключен в режиме:", appEnv)
	}

	// ================= AUTH ROUTES =================
	auth := r.Group("/auth")
	{
		// Public
		auth.POST("/register", handlers.RegisterHandler(db))
		auth.POST("/login", handlers.LoginHandler(db))
		auth.POST("/refresh", handlers.RefreshHandler(db))

		// OAuth
		auth.GET("/oauth/:provider", handlers.OAuthStartHandler())
		auth.GET("/oauth/:provider/callback", handlers.OAuthCallbackHandler(db))

		// Protected
		auth.GET("/whoami",
			middleware.AuthMiddleware(authService),
			handlers.WhoAmIHandler(db),
		)

		auth.POST("/logout",
			middleware.AuthMiddleware(authService),
			handlers.LogoutHandler(db),
		)

		auth.POST("/logout-all",
			middleware.AuthMiddleware(authService),
			handlers.LogoutAllHandler(db),
		)
	}

	// ================= ARTICLES ROUTES (CRUD) =================
	articles := r.Group("/api/articles")
	articles.Use(middleware.AuthMiddleware(authService))
	{
		articles.POST("/", handlers.CreateArticleHandler(db))
		articles.GET("/", handlers.GetArticlesHandler(db))
		articles.GET("/:id", handlers.GetArticleByIDHandler(db))
		articles.PUT("/:id", handlers.UpdateArticleHandler(db))
		articles.DELETE("/:id", handlers.DeleteArticleHandler(db))
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"env":    appEnv,
		})
	})

	// ================= START =================
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("🚀 Server running on port:", port)
	log.Println("📝 Health check: http://localhost:" + port + "/health")

	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("failed to start server:", err)
	}
}
