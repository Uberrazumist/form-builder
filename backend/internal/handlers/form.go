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

// ---------- SubmitResponse (обновлённый) ----------
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
        if err := db.Preload("Questions").First(&form, "id = ?", input.FormID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
            return
        }

        // Получаем userID
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
            return
        }
        uid := uuid.MustParse(userID)

        // Проверяем занятость для is_booking вопросов
        // Собираем информацию о бронированиях
        var bookingsToCreate []models.Booking

        for _, q := range form.Questions {
            if q.IsBooking && q.DictionaryID != nil {
                val, exists := input.Answers[q.ID.String()]
                if !exists {
                    if q.IsRequired {
                        c.JSON(http.StatusBadRequest, gin.H{"error": "Required booking question not answered"})
                        return
                    }
                    continue
                }

                // Получаем ID слота времени
                slotIDStr, ok := val.(string)
                if !ok {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot value format"})
                    return
                }
                slotID, err := uuid.Parse(slotIDStr)
                if err != nil {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slot ID"})
                    return
                }

                // Теперь нужно получить teacher_id и date из других ответов
                // Поскольку мы не знаем, какие вопросы содержат учителя и дату, мы можем их найти по типу вопроса или по зависимости.
                // Для простоты: предположим, что перед is_booking вопросом есть вопросы со справочником "Учителя" и "Дата".
                // Реализуем поиск по типу вопроса: если вопрос типа dictionary и справочник "Учителя" – берём его ответ.
                // Если вопрос типа date – берём его ответ.
                // Это логика для MVP.

                var teacherID uuid.UUID
                var bookingDate time.Time

                // Ищем teacher_id среди ответов
                for _, q2 := range form.Questions {
                    if q2.Type == "dictionary" && q2.DictionaryID != nil {
                        // Проверяем, что справочник – "Учителя"
                        var dict models.Dictionary
                        if err := db.First(&dict, "id = ?", q2.DictionaryID).Error; err == nil {
                            if dict.Name == "Учителя" || dict.Name == "Teachers" {
                                if val2, exists2 := input.Answers[q2.ID.String()]; exists2 {
                                    if str, ok2 := val2.(string); ok2 {
                                        teacherID = uuid.MustParse(str)
                                        break
                                    }
                                }
                            }
                        }
                    }
                }

                // Ищем дату среди ответов (вопрос типа date)
                for _, q2 := range form.Questions {
                    if q2.Type == "date" {
                        if val2, exists2 := input.Answers[q2.ID.String()]; exists2 {
                            if str, ok2 := val2.(string); ok2 {
                                if t, err := time.Parse("2006-01-02", str); err == nil {
                                    bookingDate = t
                                    break
                                }
                            }
                        }
                    }
                }

                // Если teacherID или date не найдены – ошибка
                if teacherID == uuid.Nil || bookingDate.IsZero() {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Missing teacher or date for booking"})
                    return
                }

                // Проверяем, не занят ли слот
                var count int64
                if err := db.Model(&models.Booking{}).
                    Where("teacher_id = ? AND date = ? AND slot_id = ?", teacherID, bookingDate, slotID).
                    Count(&count).Error; err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check booking"})
                    return
                }
                if count > 0 {
                    c.JSON(http.StatusConflict, gin.H{"error": "Selected slot is already booked"})
                    return
                }

                // Сохраняем бронирование для создания позже
                bookingsToCreate = append(bookingsToCreate, models.Booking{
                    FormID:    uuid.MustParse(input.FormID),
                    UserID:    uid,
                    TeacherID: teacherID,
                    Date:      bookingDate,
                    SlotID:    slotID,
                })
            }
        }

        // Сохраняем ответ
        answersJSON, _ := json.Marshal(input.Answers)
        resp := models.Response{
            FormID:  uuid.MustParse(input.FormID),
            Answers: datatypes.JSON(answersJSON),
        }
        if uid != uuid.Nil {
            resp.UserID = &uid
        }
        if err := db.Create(&resp).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save response"})
            return
        }

        // Создаём бронирования
        for _, booking := range bookingsToCreate {
            if err := db.Create(&booking).Error; err != nil {
                // Логируем ошибку, но не прерываем процесс (можно добавить в лог)
                // c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
                // return
            }
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
    ID            *uuid.UUID  `json:"id"`
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

        db.Where("form_id = ?", form.ID).Delete(&models.Response{})
        db.Where("form_id = ?", form.ID).Delete(&models.Question{})
        db.Delete(&form)

        c.JSON(http.StatusOK, gin.H{"message": "Form deleted successfully"})
    }
}
