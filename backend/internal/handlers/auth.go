package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/Uberrazumist/form-builder/backend/internal/models"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type RegisterInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    FullName string `json:"full_name"`
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func Login(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input LoginInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var user models.User
        if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
            return
        }

        if !user.Verified {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not verified. Please check your email for confirmation code."})
            return
        }

        if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
            return
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "userID": user.ID.String(),
            "email":  user.Email,
            "role":   user.Role,
            "exp":    time.Now().Add(time.Hour * 72).Unix(),
        })
        tokenString, err := token.SignedString([]byte(jwtSecret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "token": tokenString,
            "user": gin.H{
                "id":       user.ID,
                "email":    user.Email,
                "name":     user.FullName,
                "role":     user.Role,
                "verified": user.Verified,
            },
        })
    }
}
