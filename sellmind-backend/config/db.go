package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() {
	// Загружаем .env (если есть)
	godotenv.Load()

	// Собираем строку подключения
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Failed to parse pgx config:", err)
	}

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Проверим подключение
	err = DB.Ping(context.Background())
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to PostgreSQL at", os.Getenv("DB_NAME"))
}