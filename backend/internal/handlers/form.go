package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"

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

// ---------- GetForm (с правильным маппингом поля Questions) ----------
func GetForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        formID := c.Param("id")
        userID := c.GetString("userID")

        var form models.Form
        if err := db.Preload("Questions").First(&form, "id = ?", formID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
            return
        }

        // Проверка доступа
        if userID != "" && form.CreatedBy.String() == userID {
            // владелец
        } else if form.IsPublic {
            // публичная
        } else {
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
            return
        }

        // Формируем ответ с ключом "Questions" (с заглавной)
        response := gin.H{
            "ID":          form.ID,
            "Title":       form.Title,
            "Description": form.Description,
            "CreatedBy":   form.CreatedBy,
            "CreatedAt":   form.CreatedAt,
            "UpdatedAt":   form.UpdatedAt,
            "IsPublished": form.IsPublished,
            "IsPublic":    form.IsPublic,
            "Settings":    form.Settings,
            "Questions":   form.Questions, // здесь ключ с заглавной
        }
        c.JSON(http.StatusOK, response)
    }
}

// ---------- SubmitResponse (финальная версия) ----------
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

        // Опциональная авторизация
        userID := c.GetString("userID")
        var uid uuid.UUID
        var isAuthenticated bool
        if userID != "" {
            uid = uuid.MustParse(userID)
            isAuthenticated = true
        } else {
            if !form.IsPublic {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Для заполнения этой формы требуется авторизация"})
                return
            }
            isAuthenticated = false
            uid = uuid.Nil
        }

        // --- Автоопределение вопросов ---
        var bookingQuestion *models.Question // IsBooking == true, тип dictionary
        var resourceQuestion *models.Question // родительский справочник для времени
        var dateQuestion *models.Question     // тип date

        for _, q := range form.Questions {
            if q.Type == "date" {
                dateQuestion = &q
                continue
            }
            if q.Type == "dictionary" && q.IsBooking {
                bookingQuestion = &q
            }
        }

        // Если нет бронирования – просто сохраняем ответ
        if bookingQuestion == nil {
            answersJSON, _ := json.Marshal(input.Answers)
            resp := models.Response{
                FormID:  uuid.MustParse(input.FormID),
                Answers: datatypes.JSON(answersJSON),
            }
            if isAuthenticated {
                resp.UserID = &uid
            }
            if err := db.Create(&resp).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить ответ"})
                return
            }
            c.JSON(http.StatusCreated, gin.H{"message": "Ответ сохранён"})
            return
        }

        // --- Извлечение slot_id ---
        slotVal, ok := input.Answers[bookingQuestion.ID.String()]
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Значение слота отсутствует"})
            return
        }
        slotIDStr, ok := slotVal.(string)
        if !ok || slotIDStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Пожалуйста, выберите корректные значения в справочниках"})
            return
        }
        slotID, err := uuid.Parse(slotIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный UUID слота"})
            return
        }

        // --- Определение ресурса через выбранный слот ---
        var timeDictItem models.DictionaryItem
        if err := db.First(&timeDictItem, "id = ?", slotID).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Выбранный слот не найден в справочнике времени"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении элемента справочника"})
            }
            return
        }
        if timeDictItem.ParentID == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Слот времени должен быть привязан к ресурсу"})
            return
        }
        var parentItem models.DictionaryItem
        if err := db.First(&parentItem, "id = ?", timeDictItem.ParentID).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении родительского элемента"})
            return
        }
        resourceDictionaryID := parentItem.DictionaryID

        // Ищем вопрос с этим справочником
        for _, q := range form.Questions {
            if q.Type == "dictionary" && q.DictionaryID != nil && *q.DictionaryID == resourceDictionaryID {
                resourceQuestion = &q
                break
            }
        }
        if resourceQuestion == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Вопрос ресурса не найден в конфигурации формы"})
            return
        }

        // Проверяем наличие даты
        if dateQuestion == nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Вопрос даты не найден в конфигурации формы"})
            return
        }

        // --- Извлечение resource_id ---
        resourceVal, ok := input.Answers[resourceQuestion.ID.String()]
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Значение ресурса отсутствует"})
            return
        }
        resourceIDStr, ok := resourceVal.(string)
        if !ok || resourceIDStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Пожалуйста, выберите корректные значения в справочниках"})
            return
        }
        resourceID, err := uuid.Parse(resourceIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный UUID ресурса"})
            return
        }

        // --- Извлечение даты ---
        dateVal, ok := input.Answers[dateQuestion.ID.String()]
        if !ok {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Значение даты отсутствует"})
            return
        }
        dateStr, ok := dateVal.(string)
        if !ok || dateStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Пожалуйста, выберите корректную дату"})
            return
        }
        bookingDate, err := time.Parse("2006-01-02", dateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты, используйте YYYY-MM-DD"})
            return
        }

        // --- Транзакция с FOR UPDATE ---
        tx := db.Begin()
        if tx.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось начать транзакцию"})
            return
        }

        var existingBooking models.Booking
        err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            Where("teacher_id = ? AND slot_id = ? AND date = ?", resourceID, slotID, bookingDate).
            First(&existingBooking).Error

        if err == nil {
            tx.Rollback()
            c.JSON(http.StatusConflict, gin.H{"error": "Выбранное время уже занято, пожалуйста, выберите другой слот"})
            return
        } else if err != gorm.ErrRecordNotFound {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки доступности слота"})
            return
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
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать бронирование"})
            return
        }

        // Сохраняем ответ
        answersJSON, _ := json.Marshal(input.Answers)
        resp := models.Response{
            FormID:  uuid.MustParse(input.FormID),
            Answers: datatypes.JSON(answersJSON),
        }
        if isAuthenticated {
            resp.UserID = &uid
        }
        if err := tx.Create(&resp).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить ответ"})
            return
        }

        if err := tx.Commit().Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось завершить транзакцию"})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"message": "Ответ успешно сохранён"})
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
