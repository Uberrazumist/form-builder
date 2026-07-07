package handlers

import (
    "net/http"
    "unicode"

    "github.com/gin-gonic/gin"
    "github.com/Uberrazumist/form-builder/backend/internal/models"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type RegisterInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    FullName string `json:"full_name"`
}

func Register(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input RegisterInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // --- Новая валидация пароля ---
        if len(input.Password) < 8 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
            return
        }
        hasLetter := false
        hasDigit := false
        for _, ch := range input.Password {
            if unicode.IsLetter(ch) {
                hasLetter = true
            }
            if unicode.IsDigit(ch) {
                hasDigit = true
            }
        }
        if !hasLetter || !hasDigit {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Password must contain both letters and digits"})
            return
        }
        // --- Конец валидации ---

        // Проверяем, существует ли пользователь
        var existing models.User
        if err := db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
            c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
            return
        }

        // Хешируем пароль
        hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
            return
        }

        user := models.User{
            Email:    input.Email,
            Password: string(hashed),
            FullName: input.FullName,
            Role:     "viewer",
        }
        if err := db.Create(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
            return
        }

        c.JSON(http.StatusCreated, gin.H{
            "message": "User registered successfully",
            "user": gin.H{
                "id":    user.ID,
                "email": user.Email,
                "name":  user.FullName,
                "role":  user.Role,
            },
        })
    }
}
