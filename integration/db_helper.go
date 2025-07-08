package integration

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func setupDB() (*sqlx.DB, error) {
	// Установка значений по умолчанию
	host := getEnv("DB_HOST", "localhost")
	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "postgres") // Изменено на "postgres"

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		dbname,
	)

	log.Printf("Connecting to database: %s", connStr)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Printf("Connection failed: %v", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Printf("Ping failed: %v", err)
		return nil, err
	}

	log.Println("Database connected successfully")
	return db, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
