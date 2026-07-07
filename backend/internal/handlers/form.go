package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/Uberrazumist/form-builder/backend/internal/models"
    "gorm.io/gorm"
)

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
        userID := c.MustGet("userID").(string)

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
        userID := c.MustGet("userID").(string)

        var forms []models.Form
        if err := db.Where("created_by = ?", userID).Order("created_at desc").Find(&forms).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forms"})
            return
        }

        c.JSON(http.StatusOK, forms)
    }
}
