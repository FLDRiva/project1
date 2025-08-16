package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"sellmind-backend/config"
	"sellmind-backend/handlers"
)

func main() {
	// Загружаем .env
	godotenv.Load()

	// Подключаемся к БД
	config.ConnectDB()
	defer config.DB.Close()
	
	// Режим Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Проверка запуска бэка
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"db":     "connected",
			"app":    "sellmind-backend",
		})
	})

	// Авторизация
	r.POST("/api/auth", handlers.AuthHandler)

	// Роут в AI Qwen
	r.POST("/api/ai/generate-response", handlers.GenerateResponseHandler)

	// Порт из .env
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3001" 
	}

	log.Printf("Server is running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}