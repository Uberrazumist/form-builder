package handlers

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

// CreateForm – POST /api/forms
func CreateForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDStr := c.GetString("userID")
        if userIDStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор пользователя"})
            return
        }

        var input struct {
            Title       string                 `json:"title" binding:"required"`
            Description string                 `json:"description"`
            IsPublished bool                   `json:"is_published"`
            IsPublic    bool                   `json:"is_public"`
            Settings    map[string]interface{} `json:"settings"`
            Questions   []struct {
                Type         string                 `json:"type" binding:"required"`
                Title        string                 `json:"title" binding:"required"`
                Description  string                 `json:"description"`
                OrderIndex   int                    `json:"order_index"`
                IsRequired   bool                   `json:"is_required"`
                Options      []interface{}          `json:"options"`
                Validation   map[string]interface{} `json:"validation"`
                DependsOn    *string                `json:"depends_on"`
                DictionaryID *string                `json:"dictionary_id"`
                IsBooking    bool                   `json:"is_booking"`
            } `json:"questions"`
        }

        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON: " + err.Error()})
            return
        }

        settingsJSON, err := json.Marshal(input.Settings)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации настроек"})
            return
        }

        tx := db.Begin()
        form := models.Form{
            Title:       input.Title,
            Description: input.Description,
            CreatedBy:   userID,
            IsPublished: input.IsPublished,
            IsPublic:    input.IsPublic,
            Settings:    datatypes.JSON(settingsJSON),
        }
        if err := tx.Create(&form).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания формы"})
            return
        }

        for i, q := range input.Questions {
            var dependsOn *uuid.UUID
            if q.DependsOn != nil && *q.DependsOn != "" {
                parsed, err := uuid.Parse(*q.DependsOn)
                if err != nil {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID родительского вопроса"})
                    return
                }
                dependsOn = &parsed
            }
            var dictID *uuid.UUID
            if q.DictionaryID != nil && *q.DictionaryID != "" {
                parsed, err := uuid.Parse(*q.DictionaryID)
                if err != nil {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID справочника"})
                    return
                }
                dictID = &parsed
            }

            optionsJSON, err := json.Marshal(q.Options)
            if err != nil {
                tx.Rollback()
                c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации вариантов"})
                return
            }
            validationJSON, err := json.Marshal(q.Validation)
            if err != nil {
                tx.Rollback()
                c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации валидации"})
                return
            }

            question := models.Question{
                FormID:       form.ID,
                Type:         q.Type,
                Title:        q.Title,
                Description:  q.Description,
                OrderIndex:   q.OrderIndex,
                IsRequired:   q.IsRequired,
                Options:      datatypes.JSON(optionsJSON),
                Validation:   datatypes.JSON(validationJSON),
                DependsOn:    dependsOn,
                DictionaryID: dictID,
                IsBooking:    q.IsBooking,
            }
            if question.OrderIndex == 0 {
                question.OrderIndex = i
            }
            if err := tx.Create(&question).Error; err != nil {
                tx.Rollback()
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания вопросов"})
                return
            }
        }

        tx.Commit()
        c.JSON(http.StatusCreated, gin.H{
            "id":         form.ID.String(),
            "title":      form.Title,
            "created_at": form.CreatedAt,
        })
    }
}

// GetForm – GET /api/forms/:id
func GetForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        formID := c.Param("id")
        var form models.Form
        if err := db.Preload("Questions").First(&form, "id = ?", formID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Форма не найдена"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
            }
            return
        }

        userIDStr := c.GetString("userID")
        if userIDStr == "" {
            if !form.IsPublic {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Форма приватная, требуется авторизация"})
                return
            }
        } else {
            userID, err := uuid.Parse(userIDStr)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор пользователя"})
                return
            }
            if form.CreatedBy != userID && !form.IsPublic {
                c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещён"})
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{
            "id":           form.ID.String(),
            "title":        form.Title,
            "description":  form.Description,
            "created_by":   form.CreatedBy.String(),
            "created_at":   form.CreatedAt,
            "updated_at":   form.UpdatedAt,
            "is_published": form.IsPublished,
            "is_public":    form.IsPublic,
            "settings":     form.Settings,
            "questions":    form.Questions,
        })
    }
}

// ListForms – GET /api/forms
func ListForms(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDStr := c.GetString("userID")
        if userIDStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор пользователя"})
            return
        }

        var forms []models.Form
        if err := db.Where("created_by = ?", userID).Find(&forms).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка"})
            return
        }

        result := make([]gin.H, len(forms))
        for i, f := range forms {
            result[i] = gin.H{
                "id":           f.ID.String(),
                "title":        f.Title,
                "description":  f.Description,
                "created_at":   f.CreatedAt,
                "updated_at":   f.UpdatedAt,
                "is_published": f.IsPublished,
                "is_public":    f.IsPublic,
            }
        }
        c.JSON(http.StatusOK, result)
    }
}

// UpdateForm – PUT /api/forms/:id
func UpdateForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDStr := c.GetString("userID")
        if userIDStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор пользователя"})
            return
        }

        formID := c.Param("id")
        var form models.Form
        if err := db.First(&form, "id = ? AND created_by = ?", formID, userID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Форма не найдена или доступ запрещён"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
            }
            return
        }

        var input struct {
            Title       string                 `json:"title"`
            Description string                 `json:"description"`
            IsPublished bool                   `json:"is_published"`
            IsPublic    bool                   `json:"is_public"`
            Settings    map[string]interface{} `json:"settings"`
            Questions   []struct {
                ID           *string                `json:"id,omitempty"`
                Type         string                 `json:"type" binding:"required"`
                Title        string                 `json:"title" binding:"required"`
                Description  string                 `json:"description"`
                OrderIndex   int                    `json:"order_index"`
                IsRequired   bool                   `json:"is_required"`
                Options      []interface{}          `json:"options"`
                Validation   map[string]interface{} `json:"validation"`
                DependsOn    *string                `json:"depends_on"`
                DictionaryID *string                `json:"dictionary_id"`
                IsBooking    bool                   `json:"is_booking"`
            } `json:"questions"`
        }

        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON: " + err.Error()})
            return
        }

        settingsJSON, err := json.Marshal(input.Settings)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации настроек"})
            return
        }

        tx := db.Begin()
        updates := map[string]interface{}{
            "title":        input.Title,
            "description":  input.Description,
            "is_published": input.IsPublished,
            "is_public":    input.IsPublic,
            "settings":     datatypes.JSON(settingsJSON),
        }
        if err := tx.Model(&form).Updates(updates).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления формы"})
            return
        }

        if err := tx.Where("form_id = ?", form.ID).Delete(&models.Question{}).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления старых вопросов"})
            return
        }

        for i, q := range input.Questions {
            var dependsOn *uuid.UUID
            if q.DependsOn != nil && *q.DependsOn != "" {
                parsed, err := uuid.Parse(*q.DependsOn)
                if err != nil {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID родительского вопроса"})
                    return
                }
                dependsOn = &parsed
            }
            var dictID *uuid.UUID
            if q.DictionaryID != nil && *q.DictionaryID != "" {
                parsed, err := uuid.Parse(*q.DictionaryID)
                if err != nil {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID справочника"})
                    return
                }
                dictID = &parsed
            }

            optionsJSON, err := json.Marshal(q.Options)
            if err != nil {
                tx.Rollback()
                c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации вариантов"})
                return
            }
            validationJSON, err := json.Marshal(q.Validation)
            if err != nil {
                tx.Rollback()
                c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации валидации"})
                return
            }

            question := models.Question{
                FormID:       form.ID,
                Type:         q.Type,
                Title:        q.Title,
                Description:  q.Description,
                OrderIndex:   q.OrderIndex,
                IsRequired:   q.IsRequired,
                Options:      datatypes.JSON(optionsJSON),
                Validation:   datatypes.JSON(validationJSON),
                DependsOn:    dependsOn,
                DictionaryID: dictID,
                IsBooking:    q.IsBooking,
            }
            if question.OrderIndex == 0 {
                question.OrderIndex = i
            }
            if err := tx.Create(&question).Error; err != nil {
                tx.Rollback()
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания вопросов"})
                return
            }
        }

        tx.Commit()
        c.JSON(http.StatusOK, gin.H{"message": "Форма обновлена"})
    }
}

// DeleteForm – DELETE /api/forms/:id
func DeleteForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDStr := c.GetString("userID")
        if userIDStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор пользователя"})
            return
        }

        formID := c.Param("id")
        result := db.Where("id = ? AND created_by = ?", formID, userID).Delete(&models.Form{})
        if result.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
            return
        }
        if result.RowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "Форма не найдена или доступ запрещён"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Форма удалена"})
    }
}

// SubmitResponse – POST /api/responses
func SubmitResponse(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDStr := c.GetString("userID")

        var input struct {
            FormID  string                 `json:"form_id" binding:"required"`
            Answers map[string]interface{} `json:"answers" binding:"required"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON: " + err.Error()})
            return
        }

        var form models.Form
        if err := db.Preload("Questions").First(&form, "id = ?", input.FormID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Форма не найдена"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
            }
            return
        }

        if userIDStr == "" && !form.IsPublic {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Форма приватная, требуется авторизация"})
            return
        }

        // Валидация обязательных вопросов с корректной обработкой строк и массивов
        for _, q := range form.Questions {
            if q.IsRequired {
                val, has := input.Answers[q.ID.String()]
                if !has || val == nil {
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Вопрос '%s' обязателен", q.Title)})
                    return
                }
                // Проверка на пустую строку для текстовых полей
                if str, ok := val.(string); ok && str == "" {
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Вопрос '%s' обязателен", q.Title)})
                    return
                }
                // Проверка на пустой массив для чекбоксов
                if arr, ok := val.([]interface{}); ok && len(arr) == 0 {
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Вопрос '%s' обязателен", q.Title)})
                    return
                }
            }
        }

        // Поиск вопросов типа "schedule" (новый стиль бронирования)
        var scheduleQuestions []models.Question
        for i := range form.Questions {
        // Поиск вопросов типа "schedule" (новый стиль бронирования)
        var scheduleQuestions []models.Question
        for i := range form.Questions {
            q := &form.Questions[i]
            if q.Type == "schedule" {
                scheduleQuestions = append(scheduleQuestions, *q)
            }
        }

        // Обработка вопросов типа "schedule" (новый стиль бронирования)
        if len(scheduleQuestions) > 0 {
            tx := db.Begin()

            for _, sq := range scheduleQuestions {
                answerRaw, has := input.Answers[sq.ID.String()]
                if !has || answerRaw == nil {
                    // Проверяем обязательность
                    if sq.IsRequired {
                        tx.Rollback()
                        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Вопрос '%s' обязателен", sq.Title)})
                        return
                    }
                    continue
                }

                // answerRaw — map: {"date": "2026-07-20", "start_time": "09:00:00Z", "end_time": "09:45:00Z"}
                answerMap, ok := answerRaw.(map[string]interface{})
                if !ok {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Неверный формат ответа на вопрос '%s'", sq.Title)})
                    return
                }

                dateStr, ok := answerMap["date"].(string)
                if !ok || dateStr == "" {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Не указана дата в вопросе '%s'", sq.Title)})
                    return
                }

                startTimeStr, ok := answerMap["start_time"].(string)
                endTimeStr, ok2 := answerMap["end_time"].(string)
                if !ok || !ok2 || startTimeStr == "" || endTimeStr == "" {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Не указаны время в вопросе '%s'", sq.Title)})
                    return
                }

                date, err := time.Parse("2006-01-02", dateStr)
                if err != nil {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Неверный формат даты в вопросе '%s'", sq.Title)})
                    return
                }

                startTime, err := time.Parse(time.RFC3339, startTimeStr)
                if err != nil {
                    // Пробуем формат "15:04"
                    t, parseErr := time.Parse("15:04", startTimeStr)
                    if parseErr != nil {
                        tx.Rollback()
                        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Неверный формат start_time в вопросе '%s'", sq.Title)})
                        return
                    }
                    startTime = time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
                }

                endTime, err := time.Parse(time.RFC3339, endTimeStr)
                if err != nil {
                    // Пробуем формат "15:04"
                    t, parseErr := time.Parse("15:04", endTimeStr)
                    if parseErr != nil {
                        tx.Rollback()
                        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Неверный формат end_time в вопросе '%s'", sq.Title)})
                        return
                    }
                    endTime = time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
                }

                // resource_id берём из depends_on (родительский вопрос графа — учитель/кабинет)
                var resourceID uuid.UUID
                if sq.DependsOn != nil {
                    resourceID = *sq.DependsOn
                } else {
                    tx.Rollback()
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Вопрос '%s' не привязан к ресурсу", sq.Title)})
                    return
                }

                // Атомарная проверка + создание бронирования
                var existingBooking models.Booking
                err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
                    Where("resource_id = ? AND date = ? AND start_time = ?", resourceID, date, startTime).
                    First(&existingBooking).Error

                if err == nil {
                    tx.Rollback()
                    c.JSON(http.StatusConflict, gin.H{"error": "Выбранное время уже занято, пожалуйста, выберите другой слот"})
                    return
                } else if !errors.Is(err, gorm.ErrRecordNotFound) {
                    tx.Rollback()
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки бронирования"})
                    return
                }

                var userID uuid.UUID
                if userIDStr != "" {
                    parsed, err := uuid.Parse(userIDStr)
                    if err != nil {
                        tx.Rollback()
                        c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
                        return
                    }
                    userID = parsed
                } else {
                    userID = uuid.Nil
                }

                booking := models.Booking{
                    FormID:     form.ID,
                    UserID:     userID,
                    ResourceID: resourceID,
                    Date:       date,
                    StartTime:  startTime,
                    EndTime:    endTime,
                }
                if err := tx.Create(&booking).Error; err != nil {
                    tx.Rollback()
                    // Проверяем unique violation
                    if err.Error() != "" {
                        c.JSON(http.StatusConflict, gin.H{"error": "Выбранное время уже занято, пожалуйста, выберите другой слот"})
                        return
                    }
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания бронирования"})
                    return
                }
            }

            // Сохраняем ответ на форму (без schedule-вопросов, они уже в Booking)
            var userIDPtr *uuid.UUID
            if userIDStr != "" {
                parsed, _ := uuid.Parse(userIDStr)
                userIDPtr = &parsed
            }

            // Удаляем schedule-ответы из answers, чтобы не дублировать
            filteredAnswers := make(map[string]interface{})
            for k, v := range input.Answers {
                filteredAnswers[k] = v
            }
            for _, sq := range scheduleQuestions {
                delete(filteredAnswers, sq.ID.String())
            }

            answersJSON, err := json.Marshal(filteredAnswers)
            if err != nil {
                tx.Rollback()
                c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации ответов"})
                return
            }
            response := models.Response{
                FormID:  form.ID,
                UserID:  userIDPtr,
                Answers: datatypes.JSON(answersJSON),
            }
            if err := tx.Create(&response).Error; err != nil {
                tx.Rollback()
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ответа"})
                return
            }

            tx.Commit()
            c.JSON(http.StatusCreated, gin.H{"message": "Ответ сохранён и бронирование выполнено"})
            return
        }

        // Если бронирования нет – просто сохраняем ответ
        var userIDPtr *uuid.UUID
        if userIDStr != "" {
            parsed, _ := uuid.Parse(userIDStr)
            userIDPtr = &parsed
        }
        answersJSON, err := json.Marshal(input.Answers)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации ответов"})
            return
        }
        response := models.Response{
            FormID:  form.ID,
            UserID:  userIDPtr,
            Answers: datatypes.JSON(answersJSON),
        }
        if err := db.Create(&response).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ответа"})
            return
        }
        c.JSON(http.StatusCreated, gin.H{"message": "Ответ сохранён"})
    }
}

// GetResponses – GET /api/forms/:id/responses
func GetResponses(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userIDStr := c.GetString("userID")
        if userIDStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор пользователя"})
            return
        }

        formID := c.Param("id")
        var form models.Form
        if err := db.First(&form, "id = ? AND created_by = ?", formID, userID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Форма не найдена или доступ запрещён"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
            }
            return
        }

        var responses []models.Response
        if err := db.Where("form_id = ?", formID).Find(&responses).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ответов"})
            return
        }

        c.JSON(http.StatusOK, responses)
    }
}
