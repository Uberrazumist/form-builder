package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/Uberrazumist/form-builder/backend/internal/handlers"
    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

func getJWTSecret() string {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET environment variable is not set")
    }
    return secret
}

func main() {
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
    if err := db.AutoMigrate(
        &models.User{},
        &models.Form{},
        &models.Question{},
        &models.Response{},
        &models.Class{},
        &models.Teacher{},
        &models.TimeSlot{},
        &models.Booking{},
        &models.FormPermission{},
    ); err != nil {
        log.Fatal("Migration failed:", err)
    }
    log.Println("Migration completed")

    r := gin.Default()

    jwtSecret := getJWTSecret()

    // Публичные маршруты
    r.GET("/ping", func(c *gin.Context) {
        c.String(http.StatusOK, "pong")
    })
    r.POST("/api/register", handlers.Register(db))
    r.POST("/api/login", handlers.Login(db, jwtSecret))

    // Защищённые маршруты (JWT)
    auth := r.Group("/api")
    auth.Use(func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
            return
        }
        if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
            tokenString = tokenString[7:]
        }
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }
        userID, ok := claims["userID"].(string)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
            return
        }
        c.Set("userID", userID)
        c.Next()
    })
    {
        auth.POST("/forms", handlers.CreateForm(db))
        auth.GET("/forms", handlers.ListForms(db))
        auth.GET("/forms/:id", handlers.GetForm(db))
        auth.POST("/responses", handlers.SubmitResponse(db))
        auth.GET("/forms/:id/responses", handlers.GetResponses(db))
    }

    log.Println("Server starting on :8080")
    r.Run(":8080")
}
