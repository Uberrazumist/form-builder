package handlers

import (
    "crypto/rand"
    "crypto/tls"
    "encoding/hex"
    "fmt"
    "net/http"
    "net/smtp"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

var (
    rateLimitMap = make(map[string]time.Time)
    rateLimitMux sync.Mutex
)

func generateCode() string {
    b := make([]byte, 4)
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
    addr := smtpHost + ":" + smtpPort

    var client *smtp.Client
    var err error

    if smtpPort == "465" {
        tlsConfig := &tls.Config{ServerName: smtpHost}
        conn, err := tls.Dial("tcp", addr, tlsConfig)
        if err != nil {
            return fmt.Errorf("TLS dial error: %v", err)
        }
        client, err = smtp.NewClient(conn, smtpHost)
        if err != nil {
            return fmt.Errorf("SMTP client error: %v", err)
        }
    } else {
        conn, err := smtp.Dial(addr)
        if err != nil {
            return fmt.Errorf("SMTP dial error: %v", err)
        }
        client = conn
        if err = client.StartTLS(&tls.Config{ServerName: smtpHost}); err != nil {
            return fmt.Errorf("StartTLS error: %v", err)
        }
    }
    defer client.Quit()

    if err = client.Auth(auth); err != nil {
        return fmt.Errorf("SMTP auth error: %v", err)
    }
    if err = client.Mail(smtpFrom); err != nil {
        return fmt.Errorf("SMTP mail error: %v", err)
    }
    if err = client.Rcpt(to); err != nil {
        return fmt.Errorf("SMTP rcpt error: %v", err)
    }

    w, err := client.Data()
    if err != nil {
        return fmt.Errorf("SMTP data error: %v", err)
    }
    msg := "From: " + smtpFrom + "\r\n" +
           "To: " + to + "\r\n" +
           "Subject: " + subject + "\r\n" +
           "\r\n" + body
    if _, err = w.Write([]byte(msg)); err != nil {
        return fmt.Errorf("SMTP write error: %v", err)
    }
    if err = w.Close(); err != nil {
        return fmt.Errorf("SMTP close error: %v", err)
    }

    fmt.Printf("[EMAIL] Sent to %s successfully\n", to)
    return nil
}

func canRequestCode(email string) bool {
    rateLimitMux.Lock()
    defer rateLimitMux.Unlock()
    last, ok := rateLimitMap[email]
    if !ok || time.Since(last) > 60*time.Second {
        rateLimitMap[email] = time.Now()
        return true
    }
    return false
}

func RegisterWithEmail(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input RegisterInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        if !strings.HasSuffix(input.Email, "@1367.ru") {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Only emails with domain @1367.ru are allowed"})
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

        if !canRequestCode(input.Email) {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "Please wait 60 seconds before requesting a new code"})
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
        body := fmt.Sprintf("Ваш код подтверждения: %s\nКод действителен 15 минут.\nЕсли письмо не пришло, проверьте папку «Спам».", code)
        if err := sendEmail(input.Email, subject, body); err != nil {
            fmt.Printf("[ERROR] Failed to send email: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email: " + err.Error()})
            return
        }

        c.JSON(http.StatusCreated, gin.H{
            "message": "Registration successful. Check your email (and spam folder) for confirmation code.",
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

        if len(input.Code) != 8 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Code must be exactly 8 characters"})
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

        if !strings.HasSuffix(input.Email, "@1367.ru") {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Only emails with domain @1367.ru are allowed"})
            return
        }

        var user models.User
        if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
            return
        }

        if !canRequestCode(input.Email) {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "Please wait 60 seconds before requesting a new code"})
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
        body := fmt.Sprintf("Ваш код для сброса пароля: %s\nКод действителен 15 минут.\nЕсли письмо не пришло, проверьте папку «Спам».", code)
        if err := sendEmail(input.Email, subject, body); err != nil {
            fmt.Printf("[ERROR] Failed to send reset email: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email: " + err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Reset code sent to your email (check spam folder)"})
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

        if len(input.Code) != 8 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Code must be exactly 8 characters"})
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

// ResendVerification – повторная отправка кода подтверждения
func ResendVerification(db *gorm.DB) gin.HandlerFunc {
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
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        if user.Verified {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Email already verified"})
            return
        }

        if !canRequestCode(input.Email) {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "Please wait 60 seconds before requesting a new code"})
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

        subject := "Подтверждение регистрации (повторная отправка)"
        body := fmt.Sprintf("Ваш новый код подтверждения: %s\nКод действителен 15 минут.\nЕсли письмо не пришло, проверьте папку «Спам».", code)
        if err := sendEmail(input.Email, subject, body); err != nil {
            fmt.Printf("[ERROR] Failed to send email: %v\n", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email: " + err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "New verification code sent to your email (check spam folder)"})
    }
}
