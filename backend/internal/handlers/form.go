package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Uberrazumist/form-builder/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
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
	Title       string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	IsPublic    bool                  `json:"is_public"`
	Questions   []CreateQuestionInput `json:"questions" binding:"required"`
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

		// Пустой JSON для Settings
		emptyJSON := datatypes.JSON([]byte("{}"))
		form := models.Form{
			Title:       input.Title,
			Description: input.Description,
			CreatedBy:   uuid.MustParse(userID),
			IsPublic:    input.IsPublic,
			Settings:    emptyJSON,
		}
		if err := db.Create(&form).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form"})
			return
		}

		for _, q := range input.Questions {
			// Сериализуем Options и DependsValues в JSON
			optsJSON, _ := json.Marshal(q.Options)
			depValsJSON, _ := json.Marshal(q.DependsValues)

			question := models.Question{
				FormID:        form.ID,
				Type:          q.Type,
				Title:         q.Title,
				Description:   q.Description,
				OrderIndex:    q.OrderIndex,
				IsRequired:    q.IsRequired,
				Options:       datatypes.JSON(optsJSON),
				DependsValues: datatypes.JSON(depValsJSON),
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

		c.JSON(http.StatusOK, forms)
	}
}

func GetForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		formID := c.Param("id")
		userID := c.GetString("userID")

		var form models.Form
		if err := db.Preload("Questions").First(&form, "id = ?", formID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
			return
		}

		if userID != "" && form.CreatedBy.String() == userID {
			// владелец
		} else if form.IsPublic {
			// публичная
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":           form.ID.String(),
			"title":        form.Title,
			"description":  form.Description,
			"is_public":    form.IsPublic,
			"is_published": form.IsPublished,
			"created_by":   form.CreatedBy.String(),
			"created_at":   form.CreatedAt,
			"updated_at":   form.UpdatedAt,
			"settings":     form.Settings,
			"questions":    form.Questions,
		})
	}
}

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

		var form models.Form
		if err := db.First(&form, "id = ?", input.FormID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
			return
		}

		answersJSON, _ := json.Marshal(input.Answers)
		resp := models.Response{
			FormID:  uuid.MustParse(input.FormID),
			Answers: datatypes.JSON(answersJSON),
		}

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

func GetResponses(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		formID := c.Param("id")
		userID := c.GetString("userID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var form models.Form
		if err := db.Preload("Questions").First(&form, "id = ?", formID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
			return
		}
		if form.CreatedBy.String() != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can view responses"})
			return
		}

		var responses []models.Response
		if err := db.Where("form_id = ?", formID).Order("created_at DESC").Find(&responses).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch responses"})
			return
		}

		// Маппинг: question_id (UUID) -> question type
		questionTypes := make(map[string]string)
		for _, q := range form.Questions {
			questionTypes[q.ID.String()] = q.Type
		}

		// ГЛОБАЛЬНЫЙ кэш: itemID -> Name
		dictNameCache := make(map[string]string)

		for i := range responses {
			var answers map[string]interface{}
			if len(responses[i].Answers) > 0 {
				json.Unmarshal(responses[i].Answers, &answers)
			}
			if answers == nil {
				answers = make(map[string]interface{})
			}

			for qKey, answer := range answers {
				var qIndex int

				// Вариант 1: ключ — числовая строка (индекс)
				if idx, err := strconv.Atoi(qKey); err == nil {
					qIndex = idx
				} else {
					// Вариант 2: ключ — UUID вопроса
					qType := questionTypes[qKey]
					if qType == "" {
						continue
					}

					switch qType {
					case "dictionary":
						if itemID, ok := answer.(string); ok && len(itemID) == 36 {
							if name, exists := dictNameCache[itemID]; exists {
								answers[qKey] = name
							} else {
								var dictItem models.DictionaryItem
								if err := db.Select("name").Where("id = ?", itemID).First(&dictItem).Error; err == nil {
									dictNameCache[itemID] = dictItem.Name
									answers[qKey] = dictItem.Name
								} else {
									answers[qKey] = "Не найдено"
								}
							}
						}
					case "schedule":
						if ansMap, ok := answer.(map[string]interface{}); ok {
							date, _ := ansMap["date"].(string)
							startTime, _ := ansMap["start_time"].(string)
							endTime, _ := ansMap["end_time"].(string)
							if len(startTime) > 5 {
								startTime = startTime[:5]
							}
							if len(endTime) > 5 {
								endTime = endTime[:5]
							}
							answers[qKey] = fmt.Sprintf("%s, %s–%s", date, startTime, endTime)
						}
					}
					continue
				}

				if qIndex < 0 || qIndex >= len(form.Questions) {
					continue
				}

				question := form.Questions[qIndex]

				switch question.Type {
				case "dictionary":
					if itemID, ok := answer.(string); ok && len(itemID) == 36 {
						if name, exists := dictNameCache[itemID]; exists {
							answers[qKey] = name
						} else {
							var dictItem models.DictionaryItem
							if err := db.Select("name").Where("id = ?", itemID).First(&dictItem).Error; err == nil {
								dictNameCache[itemID] = dictItem.Name
								answers[qKey] = dictItem.Name
							} else {
								answers[qKey] = "Не найдено"
							}
						}
					}
				case "schedule":
					if ansMap, ok := answer.(map[string]interface{}); ok {
						date, _ := ansMap["date"].(string)
						startTime, _ := ansMap["start_time"].(string)
						endTime, _ := ansMap["end_time"].(string)
						if len(startTime) > 5 {
							startTime = startTime[:5]
						}
						if len(endTime) > 5 {
							endTime = endTime[:5]
						}
						answers[qKey] = fmt.Sprintf("%s, %s–%s", date, startTime, endTime)
					}
				}
			}

			responses[i].Answers, _ = json.Marshal(answers)
		}

		c.JSON(http.StatusOK, gin.H{
			"responses": responses,
			"_debug": gin.H{
				"question_types": questionTypes,
				"num_questions":  len(form.Questions),
			},
		})
	}
}

type UpdateQuestionInput struct {
	ID            *uuid.UUID `json:"id"`
	Type          string     `json:"type" binding:"required"`
	Title         string     `json:"title" binding:"required"`
	Description   string     `json:"description"`
	OrderIndex    int        `json:"order_index"`
	IsRequired    bool       `json:"is_required"`
	Options       []string   `json:"options"`
	DependsOn     *string    `json:"depends_on"`
	DependsValues []string   `json:"depends_values"`
}

type UpdateFormInput struct {
	Title       string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	IsPublic    bool                  `json:"is_public"`
	Questions   []UpdateQuestionInput `json:"questions" binding:"required"`
}

func UpdateForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		formID := c.Param("id")
		var form models.Form
		if err := db.First(&form, "id = ? AND created_by = ?", formID, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Form not found or access denied"})
			return
		}

		var input UpdateFormInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		form.Title = input.Title
		form.Description = input.Description
		form.IsPublic = input.IsPublic
		if err := db.Save(&form).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update form"})
			return
		}

		var existingQuestions []models.Question
		db.Where("form_id = ?", form.ID).Find(&existingQuestions)
		existingIDs := make(map[uuid.UUID]bool)
		for _, q := range existingQuestions {
			existingIDs[q.ID] = true
		}

		incomingIDs := make(map[uuid.UUID]bool)
		for _, qInput := range input.Questions {
			if qInput.ID != nil {
				incomingIDs[*qInput.ID] = true
			}
		}

		for _, q := range existingQuestions {
			if !incomingIDs[q.ID] {
				db.Delete(&q)
			}
		}

		for _, qInput := range input.Questions {
			var question models.Question
			if qInput.ID != nil {
				if err := db.First(&question, "id = ? AND form_id = ?", qInput.ID, form.ID).Error; err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Question with ID " + qInput.ID.String() + " not found"})
					return
				}
				question.Type = qInput.Type
				question.Title = qInput.Title
				question.Description = qInput.Description
				question.OrderIndex = qInput.OrderIndex
				question.IsRequired = qInput.IsRequired
				optsJSON, _ := json.Marshal(qInput.Options)
				depValsJSON, _ := json.Marshal(qInput.DependsValues)
				question.Options = datatypes.JSON(optsJSON)
				question.DependsValues = datatypes.JSON(depValsJSON)
				if qInput.DependsOn != nil {
					dependsUUID := uuid.MustParse(*qInput.DependsOn)
					question.DependsOn = &dependsUUID
				} else {
					question.DependsOn = nil
				}
				db.Save(&question)
			} else {
				optsJSON, _ := json.Marshal(qInput.Options)
				depValsJSON, _ := json.Marshal(qInput.DependsValues)
				newQ := models.Question{
					FormID:        form.ID,
					Type:          qInput.Type,
					Title:         qInput.Title,
					Description:   qInput.Description,
					OrderIndex:    qInput.OrderIndex,
					IsRequired:    qInput.IsRequired,
					Options:       datatypes.JSON(optsJSON),
					DependsValues: datatypes.JSON(depValsJSON),
				}
				if qInput.DependsOn != nil {
					dependsUUID := uuid.MustParse(*qInput.DependsOn)
					newQ.DependsOn = &dependsUUID
				}
				db.Create(&newQ)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Form updated successfully"})
	}
}

func DeleteForm(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		formID := c.Param("id")
		var form models.Form
		if err := db.First(&form, "id = ? AND created_by = ?", formID, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Form not found or access denied"})
			return
		}

		db.Where("form_id = ?", form.ID).Delete(&models.Response{})
		db.Where("form_id = ?", form.ID).Delete(&models.Question{})
		db.Delete(&form)

		c.JSON(http.StatusOK, gin.H{"message": "Form deleted successfully"})
	}
}
