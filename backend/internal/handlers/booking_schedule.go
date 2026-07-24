package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Uberrazumist/form-builder/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// validateScheduleConfig проверяет корректность конфигурации расписания
func validateScheduleConfig(config *models.RecurringSchedule) error {
	// SlotDuration не может быть 0 или отрицательным
	if config.SlotDuration <= 0 {
		return errors.New("Длительность слота должна быть больше 0 минут")
	}

	// Разрешаем пустые weekly_intervals, если есть fixed_slots
	hasWeekly := len(config.WeeklyIntervals) > 0
	hasFixed := len(config.FixedSlots) > 0

	if !hasWeekly && !hasFixed {
		return errors.New("Добавьте хотя бы один рабочий день или разовый слот")
	}

	// Валидируем weekly_intervals только если они есть
	if hasWeekly {
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
	}

	// Валидируем fixed_slots если они есть
	if hasFixed {
		for i, fs := range config.FixedSlots {
			if fs.Date == "" {
				return fmt.Errorf("Разовый слот #%d: укажите дату", i+1)
			}
			if fs.StartTime == "" {
				return fmt.Errorf("Разовый слот #%d: укажите время начала", i+1)
			}
		}
	}

	return nil
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

		// Проверяем, есть ли уже расписание для этого ресурса
		var existingRule models.ScheduleRule
		err = db.Where("resource_id = ? AND is_deleted = false", resourceID).First(&existingRule).Error

		if err == nil {
			// Расписание уже существует — обновляем его
			existingRule.Name = input.Name
			existingRule.Recurring = datatypes.JSON(recurringJSON)
			if err := db.Save(&existingRule).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления правила"})
				return
			}
			c.JSON(http.StatusOK, existingRule)
			return
		}

		// Расписание не найдено — создаём новое
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

		// 2. Проверяем fixed_slots (приоритет выше weekly_intervals)
		// Если на эту дату есть fixed_slots — weekly_intervals игнорируются
		var hasFixedSlots bool
		if len(config.FixedSlots) > 0 {
			for _, fs := range config.FixedSlots {
				if fs.Date == targetDateStr {
					hasFixedSlots = true
					break
				}
			}
		}

		// 3. Если нет fixed_slots — берём weekly_intervals (если нет исключения)
		if !exceptionFound && !hasFixedSlots {
			for _, day := range config.WeeklyIntervals {
				if day.DayOfWeek == dayOfWeek {
					activeIntervals = day.Intervals
					break
				}
			}
		}

		if len(activeIntervals) == 0 && !hasFixedSlots {
			c.JSON(http.StatusOK, gin.H{"slots": []interface{}{}})
			return
		}

		// 4. Генерируем слоты из интервалов (weekly_intervals / exceptions)
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

		// 4.1 Сортируем сгенерированные слоты
		for i := 0; i < len(allSlots); i++ {
			for j := i + 1; j < len(allSlots); j++ {
				labelI := allSlots[i]["start_label"].(string)
				labelJ := allSlots[j]["start_label"].(string)
				if labelJ < labelI {
					allSlots[i], allSlots[j] = allSlots[j], allSlots[i]
				}
			}
		}

		// 5. Добавляем fixed_slots (только если на эту дату есть)
		if hasFixedSlots {
			for _, fs := range config.FixedSlots {
				if fs.Date == targetDateStr {
					startParts := strings.Split(fs.StartTime, ":")
					sh, _ := strconv.Atoi(startParts[0])
					sm, _ := strconv.Atoi(startParts[1])

					start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), sh, sm, 0, 0, time.UTC)
					// Автоматически вычисляем end_time из slot_duration
					end := start.Add(time.Duration(config.SlotDuration) * time.Minute)

					allSlots = append(allSlots, gin.H{
						"start_time":  start.UTC().Format(time.RFC3339),
						"end_time":    end.UTC().Format(time.RFC3339),
						"start_label": start.Format("15:04"),
						"end_label":   end.Format("15:04"),
					})
				}
			}
		}

		// 5.1 Финальная сортировка всех слотов
		for i := 0; i < len(allSlots); i++ {
			for j := i + 1; j < len(allSlots); j++ {
				labelI := allSlots[i]["start_label"].(string)
				labelJ := allSlots[j]["start_label"].(string)
				if labelJ < labelI {
					allSlots[i], allSlots[j] = allSlots[j], allSlots[i]
				}
			}
		}

		// 5.2 Дедупликация — убираем дубли по start_label
		seenLabels := make(map[string]bool)
		var dedupedSlots []gin.H
		for _, slot := range allSlots {
			label := slot["start_label"].(string)
			if !seenLabels[label] {
				seenLabels[label] = true
				dedupedSlots = append(dedupedSlots, slot)
			}
		}
		allSlots = dedupedSlots

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

// CreateBooking – POST /api/bookings
func CreateBooking(db *gorm.DB) gin.HandlerFunc {
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
			FormID     string    `json:"form_id" binding:"required"`
			ResourceID string    `json:"resource_id" binding:"required"`
			Date       time.Time `json:"date" binding:"required"`
			StartTime  time.Time `json:"start_time" binding:"required"`
			EndTime    time.Time `json:"end_time" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные бронирования"})
			return
		}

		resourceID, err := uuid.Parse(input.ResourceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID ресурса"})
			return
		}

		// Проверяем что start < end
		if !input.StartTime.Before(input.EndTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Время начала должно быть раньше окончания"})
			return
		}

		// Проверяем что бронирование укладывается в расписание
		var rule models.ScheduleRule
		if err := db.Where("resource_id = ? AND is_deleted = false", resourceID).First(&rule).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Для ресурса не настроено расписание"})
			return
		}

		var config models.RecurringSchedule
		if err := json.Unmarshal(rule.Recurring, &config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга правил расписания"})
			return
		}

		if err := validateBookingInterval(&config, input.Date, input.StartTime, input.EndTime); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Транзакция для защиты от двойного бронирования
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Проверяем что слот не занят (с блокировкой строки)
		var existingBooking models.Booking
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where(
			"resource_id = ? AND date = ? AND start_time = ?",
			resourceID, input.Date, input.StartTime,
		).First(&existingBooking).Error; err == nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Этот слот уже занят"})
			return
		}

		booking := models.Booking{
			FormID:     uuid.MustParse(input.FormID),
			UserID:     userID,
			ResourceID: resourceID,
			Date:       input.Date,
			StartTime:  input.StartTime,
			EndTime:    input.EndTime,
		}

		if err := tx.Create(&booking).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания бронирования"})
			return
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения бронирования"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":          booking.ID.String(),
			"form_id":     booking.FormID.String(),
			"resource_id": booking.ResourceID.String(),
			"date":        booking.Date.Format("2006-01-02"),
			"start_time":  booking.StartTime.Format(time.RFC3339),
			"end_time":    booking.EndTime.Format(time.RFC3339),
		})
	}
}

// validateBookingInterval проверяет, что бронирование укладывается в разрешённые интервалы расписания
// Приоритет проверки: Exceptions > FixedSlots > WeeklyIntervals
func validateBookingInterval(config *models.RecurringSchedule, date time.Time, startTime, endTime time.Time) error {
	targetDateStr := date.Format("2006-01-02")

	// 1. Проверяем исключения (приоритет выше всего)
	for _, exc := range config.Exceptions {
		if exc.Date == targetDateStr {
			if !exc.IsWorking {
				return fmt.Errorf("Выбранная дата является выходным днём")
			}
			// Если это рабочий день с особым графиком, проверяем интервалы исключения
			for _, interval := range exc.Intervals {
				if isTimeInRange(startTime, endTime, interval, date) {
					return nil
				}
			}
			return fmt.Errorf("Выбранное время не попадает в расписание на эту дату")
		}
	}

	// 2. Проверяем фиксированные слоты (приоритет выше weekly_intervals)
	for _, fs := range config.FixedSlots {
		if fs.Date == targetDateStr {
			startParts := strings.Split(fs.StartTime, ":")
			sh, _ := strconv.Atoi(startParts[0])
			sm, _ := strconv.Atoi(startParts[1])
			slotStart := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
			slotEnd := slotStart.Add(time.Duration(config.SlotDuration) * time.Minute)
			if startTime.Equal(slotStart) && endTime.Equal(slotEnd) {
				return nil
			}
		}
	}

	// 3. Проверяем недельное расписание
	dayOfWeek := int(date.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7
	}
	for _, day := range config.WeeklyIntervals {
		if day.DayOfWeek == dayOfWeek {
			for _, interval := range day.Intervals {
				if isTimeInRange(startTime, endTime, interval, date) {
					return nil
				}
			}
			return fmt.Errorf("Выбранное время не попадает в расписание")
		}
	}

	return fmt.Errorf("Выбранный день не является рабочим")
}

// isTimeInRange проверяет, что startTime-endTime укладывается в interval
func isTimeInRange(startTime, endTime time.Time, interval models.TimeInterval, date time.Time) bool {
	startParts := strings.Split(interval.Start, ":")
	endParts := strings.Split(interval.End, ":")

	sh, _ := strconv.Atoi(startParts[0])
	sm, _ := strconv.Atoi(startParts[1])
	eh, _ := strconv.Atoi(endParts[0])
	em, _ := strconv.Atoi(endParts[1])

	intervalStart := time.Date(date.Year(), date.Month(), date.Day(), sh, sm, 0, 0, time.UTC)
	intervalEnd := time.Date(date.Year(), date.Month(), date.Day(), eh, em, 0, 0, time.UTC)

	// Бронирование должно полностью укладываться в интервал
	return !startTime.Before(intervalStart) && !endTime.After(intervalEnd)
}
