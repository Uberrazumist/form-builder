package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/Uberrazumist/form-builder/backend/internal/handlers"
    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

func main() {
    // Переменные окружения
    host := os.Getenv("DB_HOST")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    port := os.Getenv("DB_PORT")
    if host == "" { host = "localhost" }
    if user == "" { user = "formuser" }
    if password == "" { password = "formpass" }
    if dbname == "" { dbname = "formsdb" }
    if port == "" { port = "5432" }

    dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=disable"

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    log.Println("Connected to database successfully")

    // Миграции
    if err := db.AutoMigrate(&models.User{}); err != nil {
        log.Fatal("Migration failed:", err)
    }
    log.Println("Migration completed")

    // Gin
    r := gin.Default()

    // Регистрация (пока только этот эндпоинт)
    r.POST("/api/register", handlers.Register(db))

    // Запуск
    log.Println("Server starting on :8080")
    r.Run(":8080")
}
