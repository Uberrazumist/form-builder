package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/Uberrazumist/form-builder/backend/internal/models"
    "gorm.io/gorm"
)

// ---------- Существующие функции (CreateForm, ListForms) ----------
type CreateQuestionInput struct {
    Type          string   `json:"type" binding:"required"`
    Title         string   `json:"title" binding:"required"`
    Description   string   `json:"description"`
    OrderIndex    int      `json:"order_index"`
    IsRequired    bool     `json:"is_required"`
    Options       []string `json:"options"`
    DependsOn     *string  `json:"depends_on"`
    DependsValues []string `json:"depends_values"`
}

type CreateFormInput struct {
    Title       string                 `json:"title" binding:"required"`
    Description string                 `json:"description"`
    IsPublic    bool                   `json:"is_public"`
    Questions   []CreateQuestionInput  `json:"questions" binding:"required"`
}

func CreateForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        var input CreateFormInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        form := models.Form{
            Title:       input.Title,
            Description: input.Description,
            CreatedBy:   uuid.MustParse(userID),
            IsPublic:    input.IsPublic,
            Settings:    make(map[string]interface{}),
        }
        if err := db.Create(&form).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form"})
            return
        }

        for _, q := range input.Questions {
            question := models.Question{
                FormID:        form.ID,
                Type:          q.Type,
                Title:         q.Title,
                Description:   q.Description,
                OrderIndex:    q.OrderIndex,
                IsRequired:    q.IsRequired,
                Options:       q.Options,
                DependsValues: q.DependsValues,
            }
            if q.DependsOn != nil {
                dependsUUID := uuid.MustParse(*q.DependsOn)
                question.DependsOn = &dependsUUID
            }
            if err := db.Create(&question).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
                return
            }
        }

        c.JSON(http.StatusCreated, gin.H{
            "id":          form.ID,
            "title":       form.Title,
            "description": form.Description,
            "is_public":   form.IsPublic,
        })
    }
}

func ListForms(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        var forms []models.Form
        if err := db.Where("created_by = ?", userID).Order("created_at desc").Find(&forms).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forms"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"forms": forms})
    }
}

// ---------- Новые функции ----------

// GetForm возвращает форму с вопросами (доступна владельцу или публичная)
func GetForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        formID := c.Param("id")
        userID := c.GetString("userID") // может быть пустым, если не авторизован

        var form models.Form
        if err := db.Preload("Questions").First(&form, "id = ?", formID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
            return
        }

        // Проверка доступа: владелец ИЛИ публичная
        if userID != "" && form.CreatedBy.String() == userID {
            // Владелец – доступ разрешён
        } else if form.IsPublic {
            // Публичная – доступ разрешён
        } else {
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
            return
        }

        c.JSON(http.StatusOK, form)
    }
}

// SubmitResponse сохраняет ответы на форму
func SubmitResponse(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            FormID  string                 `json:"form_id" binding:"required"`
            Answers map[string]interface{} `json:"answers" binding:"required"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Проверяем существование формы
        var form models.Form
        if err := db.First(&form, "id = ?", input.FormID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
            return
        }

        // Создаём ответ
        resp := models.Response{
            FormID:  uuid.MustParse(input.FormID),
            Answers: input.Answers,
        }

        // Если пользователь авторизован, привязываем его ID
        if userID, exists := c.Get("userID"); exists {
            uid := uuid.MustParse(userID.(string))
            resp.UserID = &uid
        }

        if err := db.Create(&resp).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save response"})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "Response saved successfully"})
    }
}

// GetResponses возвращает все ответы на форму (только владелец)
func GetResponses(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        formID := c.Param("id")
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

        // Проверяем, что пользователь – владелец
        var form models.Form
        if err := db.First(&form, "id = ?", formID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
            return
        }
        if form.CreatedBy.String() != userID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can view responses"})
            return
        }

        var responses []models.Response
        if err := db.Where("form_id = ?", formID).Find(&responses).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch responses"})
            return
        }

        // Добавляем информацию о пользователе (если есть)
        // Для простоты пока без preload, можно позже добавить связь
        c.JSON(http.StatusOK, gin.H{"responses": responses})
    }
}
