package handlers

import (
    "errors"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

// CreateForm – POST /api/forms
func CreateForm(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
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
                Options      map[string]interface{} `json:"options"`
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

        tx := db.Begin()
        form := models.Form{
            Title:       input.Title,
            Description: input.Description,
            CreatedBy:   userID,
            IsPublished: input.IsPublished,
            IsPublic:    input.IsPublic,
            Settings:    input.Settings,
        }
        if err := tx.Create(&form).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания формы"})
            return
        }

        for i, q := range input.Questions {
            // Если depends_on передан как пустая строка, превращаем в nil
            dependsOn := q.DependsOn
            if dependsOn != nil && *dependsOn == "" {
                dependsOn = nil
            }
            question := models.Question{
                FormID:       form.ID,
                Type:         q.Type,
                Title:        q.Title,
                Description:  q.Description,
                OrderIndex:   q.OrderIndex,
                IsRequired:   q.IsRequired,
                Options:      q.Options,
                Validation:   q.Validation,
                DependsOn:    dependsOn,
                DictionaryID: q.DictionaryID,
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
            "id":         form.ID,
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

        userID := c.GetString("userID")
        if userID == "" {
            if !form.IsPublic {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Форма приватная, требуется авторизация"})
                return
            }
        } else {
            if form.CreatedBy != userID && !form.IsPublic {
                c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещён"})
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{
            "id":           form.ID,
            "title":        form.Title,
            "description":  form.Description,
            "created_by":   form.CreatedBy,
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
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
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
                "id":           f.ID,
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
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
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
                Options      map[string]interface{} `json:"options"`
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

        tx := db.Begin()
        updates := map[string]interface{}{
            "title":        input.Title,
            "description":  input.Description,
            "is_published": input.IsPublished,
            "is_public":    input.IsPublic,
            "settings":     input.Settings,
        }
        if err := tx.Model(&form).Updates(updates).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления формы"})
            return
        }

        // Удаляем старые вопросы и создаём новые
        if err := tx.Where("form_id = ?", form.ID).Delete(&models.Question{}).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления старых вопросов"})
            return
        }

        for i, q := range input.Questions {
            dependsOn := q.DependsOn
            if dependsOn != nil && *dependsOn == "" {
                dependsOn = nil
            }
            question := models.Question{
                FormID:       form.ID,
                Type:         q.Type,
                Title:        q.Title,
                Description:  q.Description,
                OrderIndex:   q.OrderIndex,
                IsRequired:   q.IsRequired,
                Options:      q.Options,
                Validation:   q.Validation,
                DependsOn:    dependsOn,
                DictionaryID: q.DictionaryID,
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
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
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
        userID := c.GetString("userID")

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

        // Проверка прав: если пользователь не авторизован, форма должна быть публичной
        if userID == "" && !form.IsPublic {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Форма приватная, требуется авторизация"})
            return
        }

        // Валидация только обязательных вопросов (без проверки depends_on)
        for _, q := range form.Questions {
            if q.IsRequired {
                val, has := input.Answers[q.ID]
                if !has || val == nil || val == "" {
                    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Вопрос '%s' обязателен", q.Title)})
                    return
                }
                // Для чекбоксов дополнительно проверяем, что массив не пуст
                if q.Type == "checkbox" {
                    if arr, ok := val.([]interface{}); ok && len(arr) == 0 {
                        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Вопрос '%s' обязателен", q.Title)})
                        return
                    }
                }
            }
        }

        // Поиск вопросов для бронирования
        var bookingQuestion *models.Question
        var resourceQuestion *models.Question
        var dateQuestion *models.Question

        for i := range form.Questions {
            q := &form.Questions[i]
            if q.IsBooking && q.DictionaryID != nil {
                bookingQuestion = q
            }
            if q.Type == "date" {
                dateQuestion = q
            }
            if q.DictionaryID != nil && !q.IsBooking {
                resourceQuestion = q
            }
        }

        // Если есть бронирование – выполняем атомарную блокировку
        if bookingQuestion != nil {
            if dateQuestion == nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Для бронирования требуется вопрос типа 'date'"})
                return
            }
            dateStr, ok := input.Answers[dateQuestion.ID].(string)
            if !ok || dateStr == "" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Дата не указана или имеет неверный формат"})
                return
            }
            date, err := time.Parse("2006-01-02", dateStr)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты, ожидается YYYY-MM-DD"})
                return
            }

            slotIDStr, ok := input.Answers[bookingQuestion.ID].(string)
            if !ok || slotIDStr == "" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Не выбран слот для бронирования"})
                return
            }
            slotUUID, err := uuid.Parse(slotIDStr)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID слота"})
                return
            }

            var slotItem models.DictionaryItem
            if err := db.First(&slotItem, "id = ?", slotUUID).Error; err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Выбранный слот не найден"})
                return
            }
            if slotItem.ParentID == nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Слот не привязан к ресурсу"})
                return
            }
            resourceID := *slotItem.ParentID

            // Транзакция с FOR UPDATE
            tx := db.Begin()
            var existingBooking models.Booking
            err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
                Where("teacher_id = ? AND slot_id = ? AND date = ?", resourceID, slotUUID, date).
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

            // Создаём бронирование
            booking := models.Booking{
                FormID:    form.ID,
                UserID:    userID,
                TeacherID: resourceID,
                SlotID:    slotUUID.String(),
                Date:      date,
            }
            if userID == "" {
                booking.UserID = uuid.Nil.String()
            }
            if err := tx.Create(&booking).Error; err != nil {
                tx.Rollback()
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания бронирования"})
                return
            }

            // Сохраняем ответ
            var userIDPtr *string
            if userID != "" {
                userIDPtr = &userID
            }
            response := models.Response{
                FormID:  form.ID,
                UserID:  userIDPtr,
                Answers: input.Answers,
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
        var userIDPtr *string
        if userID != "" {
            userIDPtr = &userID
        }
        response := models.Response{
            FormID:  form.ID,
            UserID:  userIDPtr,
            Answers: input.Answers,
        }
        if err := db.Create(&response).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения ответа"})
            return
        }
        c.JSON(http.StatusCreated, gin.H{"message": "Ответ сохранён"})
    }
}

// GetResponses – GET /api/forms/:id/responses (только владелец)
func GetResponses(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        if userID == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
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
