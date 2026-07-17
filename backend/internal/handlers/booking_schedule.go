package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/Uberrazumist/form-builder/backend/internal/models"
)

// validateScheduleConfig проверяет корректность конфигурации расписания
func validateScheduleConfig(config *models.RecurringSchedule) error {
	// SlotDuration не может быть 0 или отрицательным
	if config.SlotDuration <= 0 {
		return errors.New("Длительность слота должна быть больше 0 минут")
	}

	if len(config.WeeklyIntervals) == 0 {
		return errors.New("Добавьте хотя бы один рабочий день")
	}

	// Проверяем каждый день недели
	for _, day := range config.WeeklyIntervals {
		if day.DayOfWeek < 1 || day.DayOfWeek > 7 {
			return fmt.Errorf("Неверный день недели: %d (ожидалось 1-7)", day.DayOfWeek)
		}

		if len(day.Intervals) == 0 {
			return fmt.Errorf("Для дня %d добавьте хотя бы один временной интервал", day.DayOfWeek)
		}

		// Сортируем интервалы по времени начала
		for i := 0; i < len(day.Intervals); i++ {
			for j := i + 1; j < len(day.Intervals); j++ {
				if day.Intervals[j].Start < day.Intervals[i].Start {
					day.Intervals[i], day.Intervals[j] = day.Intervals[j], day.Intervals[i]
				}
			}
		}

		// Проверяем каждый интервал
		for i, interval := range day.Intervals {
			if interval.Start == "" || interval.End == "" {
				return fmt.Errorf("Интервал #%d дня %d: укажите start и end", i+1, day.DayOfWeek)
			}
			if interval.Start >= interval.End {
				return fmt.Errorf("День %d, интервал #%d: время начала должно быть раньше окончания (%s < %s)",
					day.DayOfWeek, i+1, interval.Start, interval.End)
			}

			// Проверяем пересечение с следующим интервалом
			if i < len(day.Intervals)-1 {
				if interval.End > day.Intervals[i+1].Start {
					return fmt.Errorf("День %d: интервалы пересекаются (%s-%s и %s-%s)",
						day.DayOfWeek, interval.Start, interval.End,
						day.Intervals[i+1].Start, day.Intervals[i+1].End)
				}
			}
		}
	}

	return nil
}

// validateBookingInterval проверяет, что запрашиваемое бронирование укладывается
// в один из разрешённых интервалов расписания на указанную дату
func validateBookingInterval(config *models.RecurringSchedule, date time.Time, startTime, endTime time.Time) error {
	targetDateStr := date.Format("2006-01-02")
	dayOfWeek := int(date.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}

	// Определяем активные интервалы для этой даты
	var activeIntervals []models.TimeInterval

	// 1. Проверяем исключения
	for _, exc := range config.Exceptions {
		if exc.Date == targetDateStr {
			if !exc.IsWorking {
				return errors.New("На эту дату расписание не действует (выходной)")
			}
			activeIntervals = exc.Intervals
			break
		}
	}

	// 2. Если исключений нет, берём недельное расписание
	if len(activeIntervals) == 0 {
		for _, day := range config.WeeklyIntervals {
			if day.DayOfWeek == dayOfWeek {
				activeIntervals = day.Intervals
				break
			}
		}
	}

	if len(activeIntervals) == 0 {
		return errors.New("В этот день расписание не действует")
	}

	// Конвертируем время бронирования в минуты от начала дня
	bookStart := startTime.Hour()*60 + startTime.Minute()
	bookEnd := endTime.Hour()*60 + endTime.Minute()

	// Проверяем, что всё укладывается в один интервал
	for _, interval := range activeIntervals {
		intParts := strings.Split(interval.Start, ":")
		endParts := strings.Split(interval.End, ":")
		intStartH, _ := strconv.Atoi(intParts[0])
		intStartM, _ := strconv.Atoi(intParts[1])
		intEndH, _ := strconv.Atoi(endParts[0])
		intEndM, _ := strconv.Atoi(endParts[1])
		intStart := intStartH*60 + intStartM
		intEnd := intEndH*60 + intEndM

		if bookStart >= intStart && bookEnd <= intEnd {
			return nil // Всё OK — бронирование укладывается
		}
	}

	return errors.New("Выбранное время не укладывается в разрешённые интервалы расписания")
}

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
				"id":          r.ID.String(),
				"resource_id": r.ResourceID.String(),
				"name":        r.Name,
				"recurring":   r.Recurring,
				"is_deleted":  r.IsDeleted,
				"created_at":  r.CreatedAt,
				"updated_at":  r.UpdatedAt,
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

		// Парсим и валидируем recurring
		recurringJSON, err := json.Marshal(input.Recurring)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка сериализации правил"})
			return
		}

		var config models.RecurringSchedule
		if err := json.Unmarshal(recurringJSON, &config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка парсинга правил расписания"})
			return
		}
		if err := validateScheduleConfig(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			var config models.RecurringSchedule
			if err := json.Unmarshal(recurringJSON, &config); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка парсинга правил расписания"})
				return
			}
			if err := validateScheduleConfig(&config); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты"})
			return
		}

		var rule models.ScheduleRule
		if err := db.Where("resource_id = ? AND is_deleted = false", resourceID).First(&rule).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}})
			return
		}

		var config models.RecurringSchedule
		if err := json.Unmarshal(rule.Recurring, &config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга правил"})
			return
		}

		targetDateStr := targetDate.Format("2006-01-02")
		dayOfWeek := int(targetDate.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}

		var activeIntervals []models.TimeInterval

		// 1. Проверяем исключения (приоритет выше недельного расписания)
		exceptionFound := false
		for _, exc := range config.Exceptions {
			if exc.Date == targetDateStr {
				exceptionFound = true
				if !exc.IsWorking {
					c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}}) // Полный выходной
					return
				}
				activeIntervals = exc.Intervals // Используем интервалы исключения
				break
			}
		}

		// 2. Если исключения нет, берем недельное расписание
		if !exceptionFound {
			for _, day := range config.WeeklyIntervals {
				if day.DayOfWeek == dayOfWeek {
					activeIntervals = day.Intervals
					break
				}
			}
		}

		if len(activeIntervals) == 0 {
			c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}})
			return
		}

		// 3. Генерируем слоты для КАЖДОГО интервала этого дня
		var allSlots []gin.H
		slotDuration := time.Duration(config.SlotDuration) * time.Minute
		stepDuration := time.Duration(config.SlotDuration+config.BreakBetween) * time.Minute

		for _, interval := range activeIntervals {
			startParts := strings.Split(interval.Start, ":")
			endParts := strings.Split(interval.End, ":")

			sh, _ := strconv.Atoi(startParts[0])
			sm, _ := strconv.Atoi(startParts[1])
			eh, _ := strconv.Atoi(endParts[0])
			em, _ := strconv.Atoi(endParts[1])

			current := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), sh, sm, 0, 0, time.UTC)
			endTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), eh, em, 0, 0, time.UTC)

			for !current.Add(slotDuration).After(endTime) {
				slotEnd := current.Add(slotDuration)
				allSlots = append(allSlots, gin.H{
					"start_time":  current.UTC().Format(time.RFC3339),
					"end_time":    slotEnd.UTC().Format(time.RFC3339),
					"start_label": current.Format("15:04"),
					"end_label":   slotEnd.Format("15:04"),
				})
				current = current.Add(stepDuration)
			}
		}

		// 3.1 Сортируем слоты по времени (интервалы могут идти не по порядку)
		for i := 0; i < len(allSlots); i++ {
			for j := i + 1; j < len(allSlots); j++ {
				labelI := allSlots[i]["start_label"].(string)
				labelJ := allSlots[j]["start_label"].(string)
				if labelJ < labelI {
					allSlots[i], allSlots[j] = allSlots[j], allSlots[i]
				}
			}
		}

		// 4. Фильтрация занятых слотов
		var bookings []models.Booking
		db.Where("resource_id = ? AND date = ?", resourceID, targetDate).Find(&bookings)

		bookedMap := make(map[string]bool)
		for _, b := range bookings {
			bookedMap[b.StartTime.Format("15:04")] = true
		}

		var available []gin.H
		for _, slot := range allSlots {
			if !bookedMap[slot["start_label"].(string)] {
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