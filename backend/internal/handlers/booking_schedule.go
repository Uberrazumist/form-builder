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

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

// helper: парсинг времени "HH:MM" → часы, минуты
func parseTime(t string) (int, int) {
    var h, m int
    fmt.Sscanf(t, "%d:%d", &h, &m)
    if h < 0 || h > 23 || m < 0 || m > 59 {
        h, m = 9, 0
    }
    return h, m
}

// ========================
// CRUD ScheduleRule
// ========================

// ListScheduleRules – GET /api/schedules
func ListScheduleRules(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var rules []models.ScheduleRule
        if err := db.Where("is_deleted = false").Find(&rules).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки правил"})
            return
        }

        result := make([]gin.H, len(rules))
        for i, r := range rules {
            result[i] = gin.H{
                "id":         r.ID.String(),
                "resource_id": r.ResourceID.String(),
                "name":       r.Name,
                "recurring":  r.Recurring,
                "is_deleted": r.IsDeleted,
                "created_at": r.CreatedAt,
                "updated_at": r.UpdatedAt,
            }
        }
        c.JSON(http.StatusOK, result)
    }
}

// CreateScheduleRule – POST /api/schedules
func CreateScheduleRule(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            ResourceID string                 `json:"resource_id" binding:"required"`
            Name       string                 `json:"name" binding:"required"`
            Recurring  map[string]interface{} `json:"recurring" binding:"required"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON: " + err.Error()})
            return
        }

        resourceID, err := uuid.Parse(input.ResourceID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID ресурса"})
            return
        }

        recurringJSON, err := json.Marshal(input.Recurring)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации правил"})
            return
        }

        rule := models.ScheduleRule{
            ResourceID: resourceID,
            Name:       input.Name,
            Recurring:  datatypes.JSON(recurringJSON),
        }
        if err := db.Create(&rule).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания правила"})
            return
        }
        c.JSON(http.StatusCreated, rule)
    }
}

// GetScheduleRule – GET /api/schedules/:id
func GetScheduleRule(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var rule models.ScheduleRule
        if err := db.First(&rule, "id = ?", id).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Правило не найдено"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки"})
            }
            return
        }
        c.JSON(http.StatusOK, rule)
    }
}

// UpdateScheduleRule – PUT /api/schedules/:id
func UpdateScheduleRule(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var rule models.ScheduleRule
        if err := db.First(&rule, "id = ?", id).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Правило не найдено"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки"})
            }
            return
        }

        // Проверяем, есть ли бронирования для этого ресурса
        var bookingCount int64
        db.Model(&models.Booking{}).Where("resource_id = ?", rule.ResourceID).Count(&bookingCount)
        if bookingCount > 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя обновлять правило с существующими бронированиями. Создайте новое правило."})
            return
        }

        var input struct {
            ResourceID string                 `json:"resource_id"`
            Name       string                 `json:"name"`
            Recurring  map[string]interface{} `json:"recurring"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON: " + err.Error()})
            return
        }

        if input.Name != "" {
            rule.Name = input.Name
        }
        if input.ResourceID != "" {
            newResourceID, err := uuid.Parse(input.ResourceID)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID ресурса"})
                return
            }
            rule.ResourceID = newResourceID
        }
        if input.Recurring != nil {
            recurringJSON, err := json.Marshal(input.Recurring)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации правил"})
                return
            }
            rule.Recurring = datatypes.JSON(recurringJSON)
        }

        if err := db.Save(&rule).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления правила"})
            return
        }
        c.JSON(http.StatusOK, rule)
    }
}

// DeleteScheduleRule – DELETE /api/schedules/:id
func DeleteScheduleRule(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var rule models.ScheduleRule
        if err := db.First(&rule, "id = ?", id).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Правило не найдено"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки"})
            }
            return
        }

        // Проверяем бронирования
        var bookingCount int64
        db.Model(&models.Booking{}).Where("resource_id = ?", rule.ResourceID).Count(&bookingCount)
        if bookingCount > 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя удалить правило с существующими бронированиями."})
            return
        }

        rule.IsDeleted = true
        if err := db.Save(&rule).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления правила"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Правило удалено"})
    }
}

// ========================
// Генерация слотов на лету
// ========================

// GetAvailableSlots – GET /api/schedules/available?resource_id=...&date=...
func GetAvailableSlots(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        resourceIDStr := c.Query("resource_id")
        dateStr := c.Query("date")

        if resourceIDStr == "" || dateStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "resource_id и date обязательны"})
            return
        }

        resourceID, err := uuid.Parse(resourceIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID ресурса"})
            return
        }

        date, err := time.Parse("2006-01-02", dateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты, ожидается YYYY-MM-DD"})
            return
        }

        // 1. Находим активное правило расписания
        var rule models.ScheduleRule
        if err := db.Where("resource_id = ? AND is_deleted = false", resourceID).First(&rule).Error; err != nil {
            c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}})
            return
        }

        // 2. Парсим recurring
        var recurring struct {
            Type         string   `json:"type"`
            Days         []int    `json:"days"`
            StartTime    string   `json:"start_time"`
            EndTime      string   `json:"end_time"`
            SlotDuration int      `json:"slot_duration"`
            BreakBetween int      `json:"break_between"`
            StartDate    string   `json:"start_date"`
            EndDate      string   `json:"end_date"`
            Exceptions   []string `json:"exceptions"`
        }
        if err := json.Unmarshal(rule.Recurring, &recurring); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга правил"})
            return
        }

        // 3. Проверяем exceptions (блокировка отдельных дней)
        for _, exc := range recurring.Exceptions {
            if exc == date.Format("2006-01-02") {
                c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}})
                return
            }
        }

        // 4. Проверяем диапазон дат
        startDate, err1 := time.Parse("2006-01-02", recurring.StartDate)
        endDate, err2 := time.Parse("2006-01-02", recurring.EndDate)
        if err1 == nil && err2 == nil {
            if date.Before(startDate) || date.After(endDate) {
                c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}})
                return
            }
        }

        // 5. Проверяем день недели (только для weekly)
        if recurring.Type == "weekly" && len(recurring.Days) > 0 {
            dayOfWeek := int(date.Weekday())
            if dayOfWeek == 0 {
                dayOfWeek = 7
            }
            dayMatch := false
            for _, d := range recurring.Days {
                if d == dayOfWeek {
                    dayMatch = true
                    break
                }
            }
            if !dayMatch {
                c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}})
                return
            }
        }

        // 6. Генерируем слоты в памяти
        var sh, sm int
        fmt.Sscanf(recurring.StartTime, "%d:%d", &sh, &sm)
        var eh, em int
        fmt.Sscanf(recurring.EndTime, "%d:%d", &eh, &em)

        slotDuration := time.Duration(recurring.SlotDuration) * time.Minute
        breakDuration := time.Duration(recurring.BreakBetween) * time.Minute

        currentStart := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
        dayEnd := time.Date(date.Year(), date.Month(), date.Day(), eh, em, 0, 0, time.UTC)

        var allSlots []gin.H
        for {
            slotEnd := currentStart.Add(slotDuration)
            if slotEnd.After(dayEnd) {
                break
            }
            allSlots = append(allSlots, gin.H{
                "start_time": currentStart.UTC().Format(time.RFC3339),
                "end_time":   slotEnd.UTC().Format(time.RFC3339),
                "start_label": currentStart.Format("15:04"),
                "end_label":   slotEnd.Format("15:04"),
            })
            currentStart = currentStart.Add(slotDuration + breakDuration)
        }

        // 7. Получаем занятые слоты (бронирования)
        var bookings []models.Booking
        db.Where("resource_id = ? AND date = ?", resourceID, date).Find(&bookings)

        bookedStartTimes := make(map[string]bool)
        for _, b := range bookings {
            bookedStartTimes[b.StartTime.Format("15:04")] = true
        }

        // 8. Фильтруем свободные
        available := []gin.H{}
        for _, slot := range allSlots {
            startLabel := slot["start_label"].(string)
            if !bookedStartTimes[startLabel] {
                available = append(available, slot)
            }
        }

        c.JSON(http.StatusOK, gin.H{"slots": available})
    }
}

// CancelBooking – POST /api/bookings/:id/cancel
func CancelBooking(db *gorm.DB) gin.HandlerFunc {
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

        bookingID := c.Param("id")
        var booking models.Booking
        if err := db.Where("id = ? AND user_id = ?", bookingID, userID).First(&booking).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Бронирование не найдено"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка"})
            }
            return
        }

        if err := db.Delete(&booking).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отмены бронирования"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Бронирование отменено"})
    }
}
