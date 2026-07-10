package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

// ---------- Вспомогательная функция для depends_on ----------
func parseDependsOn(dep interface{}) *uuid.UUID {
    if dep == nil {
        return nil
    }
    if str, ok := dep.(string); ok && str != "" {
        if uid, err := uuid.Parse(str); err == nil {
            return &uid
        }
        return nil
    }
    return nil
}

// ---------- CreateForm ----------
type CreateQuestionInput struct {
    Type          string      `json:"type" binding:"required"`
    Title         string      `json:"title" binding:"required"`
    Description   string      `json:"description"`
    OrderIndex    int         `json:"order_index"`
    IsRequired    bool        `json:"is_required"`
    Options       []string    `json:"options"`
    DependsOn     interface{} `json:"depends_on"`
    DependsValues []string    `json:"depends_values"`
    DictionaryID  *string     `json:"dictionary_id"`
    IsBooking     bool        `json:"is_booking"`
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
            optsJSON, _ := json.Marshal(q.Options)
            depValsJSON, _ := json.Marshal(q.DependsValues)
            dependsOn := parseDependsOn(q.DependsOn)

            question := models.Question{
                FormID:        form.ID,
                Type:          q.Type,
                Title:         q.Title,
                Description:   q.Description,
                OrderIndex:    q.OrderIndex,
                IsRequired:    q.IsRequired,
                Options:       datatypes.JSON(optsJSON),
                DependsValues: datatypes.JSON(depValsJSON),
                IsBooking:     q.IsBooking,
                DependsOn:     dependsOn,
            }
            if q.DictionaryID != nil {
                dictUUID := uuid.MustParse(*q.DictionaryID)
                question.DictionaryID = &dictUUID
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

// ---------- ListForms ----------
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

// ---------- GetForm ----------
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

        c.JSON(http.StatusOK, form)
    }
}

// ---------- SubmitResponse (полностью абстрактный, автоопределение ресурсов) ----------
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

        // Получаем форму с вопросами
        var form models.Form
        if err := db.Preload("Questions").First(&form, "id = ?", input.FormID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
            return
        }

        // Получаем userID из контекста
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }
        uid := uuid.MustParse(userID)

        // --- Автоопределение вопросов ---
        var bookingQuestion *models.Question // вопрос бронирования (IsBooking == true, тип dictionary)
        var resourceQuestion *models.Question // вопрос-ресурс (любой dictionary, IsBooking == false)
        var dateQuestion *models.Question     // вопрос с типом "date"

        for _, q := range form.Questions {
            if q.Type == "date" {
                dateQuestion = &q
                continue
            }
            if q.Type == "dictionary" {
                if q.IsBooking {
                    bookingQuestion = &q
                } else {
                    // первый встречный dictionary без is_booking считаем ресурсом
                    if resourceQuestion == nil {
                        resourceQuestion = &q
                    }
                }
            }
        }

        // Если нет вопроса бронирования – просто сохраняем ответ и выходим
        if bookingQuestion == nil {
            answersJSON, _ := json.Marshal(input.Answers)
            resp := models.Response{
                FormID:  uuid.MustParse(input.FormID),
                Answers: datatypes.JSON(answersJSON),
                UserID:  &uid,
            }
            if err := db.Create(&resp).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save response"})
                return
            }
            c.JSON(http.StatusCreated, gin.H{"message": "Response saved successfully"})
            return
        }

        // Проверяем наличие вопроса-ресурса
        if resourceQuestion == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Resource question not found in form configuration"})
            return
        }

        // Извлекаем ID слота (из вопроса бронирования)
        slotVal, ok := input.Answers[bookingQuestion.ID.String()]
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Slot value missing"})
            return
        }
        slotIDStr, ok := slotVal.(string)
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Slot value must be a string (UUID)"})
            return
        }
        slotID, err := uuid.Parse(slotIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot UUID"})
            return
        }

        // Извлекаем ID ресурса (из вопроса-ресурса)
        resourceVal, ok := input.Answers[resourceQuestion.ID.String()]
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Resource value missing"})
            return
        }
        resourceIDStr, ok := resourceVal.(string)
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Resource value must be a string (UUID)"})
            return
        }
        resourceID, err := uuid.Parse(resourceIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource UUID"})
            return
        }

        // Извлекаем дату (если есть вопрос даты)
        var bookingDate time.Time
        if dateQuestion != nil {
            dateVal, ok := input.Answers[dateQuestion.ID.String()]
            if !ok {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Date value missing"})
                return
            }
            dateStr, ok := dateVal.(string)
            if !ok {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Date value must be a string"})
                return
            }
            bookingDate, err = time.Parse("2006-01-02", dateStr)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
                return
            }
        }

        // Транзакционная проверка и сохранение
        if err := db.Transaction(func(tx *gorm.DB) error {
            // Проверяем занятость
            query := tx.Model(&models.Booking{}).
                Where("teacher_id = ? AND slot_id = ?", resourceID, slotID)
            if !bookingDate.IsZero() {
                query = query.Where("date = ?", bookingDate)
            }
            var count int64
            if err := query.Count(&count).Error; err != nil {
                return err
            }
            if count > 0 {
                return gorm.ErrRecordNotFound // конфликт
            }

            // Создаём бронирование
            booking := models.Booking{
                FormID:    uuid.MustParse(input.FormID),
                UserID:    uid,
                TeacherID: resourceID,
                SlotID:    slotID,
                Date:      bookingDate,
            }
            if err := tx.Create(&booking).Error; err != nil {
                return err
            }

            // Сохраняем ответ
            answersJSON, _ := json.Marshal(input.Answers)
            resp := models.Response{
                FormID:  uuid.MustParse(input.FormID),
                Answers: datatypes.JSON(answersJSON),
                UserID:  &uid,
            }
            if err := tx.Create(&resp).Error; err != nil {
                return err
            }

            return nil
        }); err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusConflict, gin.H{"error": "Selected slot is already booked"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process booking"})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "Response saved successfully"})
    }
}

// ---------- GetResponses ----------
func GetResponses(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        formID := c.Param("id")
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }

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

        c.JSON(http.StatusOK, gin.H{"responses": responses})
    }
}

// ---------- UpdateForm ----------
type UpdateQuestionInput struct {
    ID            *uuid.UUID `json:"id"`
    Type          string     `json:"type" binding:"required"`
    Title         string     `json:"title" binding:"required"`
    Description   string     `json:"description"`
    OrderIndex    int        `json:"order_index"`
    IsRequired    bool       `json:"is_required"`
    Options       []string   `json:"options"`
    DependsOn     interface{} `json:"depends_on"`
    DependsValues []string   `json:"depends_values"`
    DictionaryID  *string    `json:"dictionary_id"`
    IsBooking     bool       `json:"is_booking"`
}

type UpdateFormInput struct {
    Title       string               `json:"title" binding:"required"`
    Description string               `json:"description"`
    IsPublic    bool                 `json:"is_public"`
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

        // Обновление вопросов
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

        // Удаляем отсутствующие
        for _, q := range existingQuestions {
            if !incomingIDs[q.ID] {
                db.Delete(&q)
            }
        }

        // Обновляем или создаём вопросы
        for _, qInput := range input.Questions {
            optsJSON, _ := json.Marshal(qInput.Options)
            depValsJSON, _ := json.Marshal(qInput.DependsValues)
            dependsOn := parseDependsOn(qInput.DependsOn)

            if qInput.ID != nil {
                var question models.Question
                if err := db.First(&question, "id = ? AND form_id = ?", qInput.ID, form.ID).Error; err != nil {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Question not found"})
                    return
                }
                question.Type = qInput.Type
                question.Title = qInput.Title
                question.Description = qInput.Description
                question.OrderIndex = qInput.OrderIndex
                question.IsRequired = qInput.IsRequired
                question.Options = datatypes.JSON(optsJSON)
                question.DependsValues = datatypes.JSON(depValsJSON)
                question.IsBooking = qInput.IsBooking
                question.DependsOn = dependsOn
                if qInput.DictionaryID != nil {
                    dictUUID := uuid.MustParse(*qInput.DictionaryID)
                    question.DictionaryID = &dictUUID
                } else {
                    question.DictionaryID = nil
                }
                db.Save(&question)
            } else {
                newQ := models.Question{
                    FormID:        form.ID,
                    Type:          qInput.Type,
                    Title:         qInput.Title,
                    Description:   qInput.Description,
                    OrderIndex:    qInput.OrderIndex,
                    IsRequired:    qInput.IsRequired,
                    Options:       datatypes.JSON(optsJSON),
                    DependsValues: datatypes.JSON(depValsJSON),
                    IsBooking:     qInput.IsBooking,
                    DependsOn:     dependsOn,
                }
                if qInput.DictionaryID != nil {
                    dictUUID := uuid.MustParse(*qInput.DictionaryID)
                    newQ.DictionaryID = &dictUUID
                }
                db.Create(&newQ)
            }
        }

        c.JSON(http.StatusOK, gin.H{"message": "Form updated successfully"})
    }
}

// ---------- DeleteForm ----------
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

        // Каскадное удаление
        db.Where("form_id = ?", form.ID).Delete(&models.Response{})
        db.Where("form_id = ?", form.ID).Delete(&models.Question{})
        db.Delete(&form)

        c.JSON(http.StatusOK, gin.H{"message": "Form deleted successfully"})
    }
}
