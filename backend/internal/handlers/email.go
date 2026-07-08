package handlers

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "net/http"
    "net/smtp"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

func generateCode() string {
    b := make([]byte, 3)
    rand.Read(b)
    return hex.EncodeToString(b)
}

func sendEmail(to, subject, body string) error {
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    smtpUser := os.Getenv("SMTP_USER")
    smtpPassword := os.Getenv("SMTP_PASSWORD")
    smtpFrom := os.Getenv("SMTP_FROM")

    if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" {
        fmt.Printf("[EMAIL] To: %s\nSubject: %s\nBody: %s\n", to, subject, body)
        return nil
    }

    auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
    msg := "From: " + smtpFrom + "\r\n" +
           "To: " + to + "\r\n" +
           "Subject: " + subject + "\r\n" +
           "\r\n" + body
    return smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpFrom, []string{to}, []byte(msg))
}

func RegisterWithEmail(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input RegisterInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        if len(input.Password) < 8 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
            return
        }
        hasLetter, hasDigit := false, false
        for _, ch := range input.Password {
            if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
                hasLetter = true
            }
            if ch >= '0' && ch <= '9' {
                hasDigit = true
            }
        }
        if !hasLetter || !hasDigit {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Password must contain letters and digits"})
            return
        }

        var existing models.User
        if err := db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
            c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
            return
        }

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
            Verified: false,
        }
        if err := db.Create(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
            return
        }

        code := generateCode()
        expiresAt := time.Now().Add(15 * time.Minute)

        verification := models.EmailVerification{
            Email:     input.Email,
            Code:      code,
            Type:      "registration",
            ExpiresAt: expiresAt,
            Used:      false,
        }
        if err := db.Create(&verification).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification code"})
            return
        }

        subject := "Подтверждение регистрации"
        body := fmt.Sprintf("Ваш код подтверждения: %s\nКод действителен 15 минут.", code)
        if err := sendEmail(input.Email, subject, body); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
            return
        }

        c.JSON(http.StatusCreated, gin.H{
            "message": "Registration successful. Check your email for confirmation code.",
            "user": gin.H{
                "id":       user.ID,
                "email":    user.Email,
                "name":     user.FullName,
                "verified": false,
            },
        })
    }
}

func VerifyEmail(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            Email string `json:"email" binding:"required,email"`
            Code  string `json:"code" binding:"required"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var verification models.EmailVerification
        if err := db.Where("email = ? AND code = ? AND type = ? AND used = ? AND expires_at > ?",
            input.Email, input.Code, "registration", false, time.Now()).
            First(&verification).Error; err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired code"})
            return
        }

        verification.Used = true
        db.Save(&verification)

        var user models.User
        if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        user.Verified = true
        db.Save(&user)

        c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
    }
}

func ForgotPassword(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            Email string `json:"email" binding:"required,email"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var user models.User
        if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
            return
        }

        code := generateCode()
        expiresAt := time.Now().Add(15 * time.Minute)

        verification := models.EmailVerification{
            Email:     input.Email,
            Code:      code,
            Type:      "reset_password",
            ExpiresAt: expiresAt,
            Used:      false,
        }
        if err := db.Create(&verification).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset code"})
            return
        }

        subject := "Сброс пароля"
        body := fmt.Sprintf("Ваш код для сброса пароля: %s\nКод действителен 15 минут.", code)
        if err := sendEmail(input.Email, subject, body); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Reset code sent to your email"})
    }
}

func ResetPassword(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            Email       string `json:"email" binding:"required,email"`
            Code        string `json:"code" binding:"required"`
            NewPassword string `json:"new_password" binding:"required,min=8"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var verification models.EmailVerification
        if err := db.Where("email = ? AND code = ? AND type = ? AND used = ? AND expires_at > ?",
            input.Email, input.Code, "reset_password", false, time.Now()).
            First(&verification).Error; err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired code"})
            return
        }

        hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
            return
        }

        var user models.User
        if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        user.Password = string(hashed)
        db.Save(&user)

        verification.Used = true
        db.Save(&verification)

        c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
    }
}
