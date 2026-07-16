package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Uberrazumist/form-builder/backend/internal/models"
)

func GetScheduleAnalytics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rules []models.ScheduleRule
		if err := db.Where("is_deleted = false").Find(&rules).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения правил"})
			return
		}

		type ResourceAnalytics struct {
			ID               string  `json:"id"`
			Name             string  `json:"name"`
			BookedSlotsWeek  int     `json:"booked_slots_week"`
			TotalSlotsWeek   int     `json:"total_slots_week"`
			OccupancyPercent float64 `json:"occupancy_percent"`
		}

		var analytics []ResourceAnalytics
		now := time.Now()
		
		// Определяем границы текущей недели (Понедельник - Воскресенье)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())
		sunday := monday.AddDate(0, 0, 6)
		sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, now.Location())

		for _, rule := range rules {
			// Получаем имя ресурса
			var item models.DictionaryItem
			db.First(&item, "id = ?", rule.ResourceID)

			// Считаем реальные бронирования на эту неделю
			var count int64
			db.Model(&models.Booking{}).
				Where("resource_id = ? AND date >= ? AND date <= ?", rule.ResourceID, monday, sunday).
				Count(&count)

			// Парсим правила для оценки общей вместимости
			var recurring struct {
				Days         []int  `json:"days"`
				SlotDuration int    `json:"slot_duration"`
				BreakBetween int    `json:"break_between"`
				StartTime    string `json:"start_time"`
				EndTime      string `json:"end_time"`
				DaysConfig   []struct {
					Day          int    `json:"day"`
					IsWorking    bool   `json:"is_working"`
					StartTime    string `json:"start_time"`
					EndTime      string `json:"end_time"`
					SlotDuration int    `json:"slot_duration"`
					BreakBetween int    `json:"break_between"`
				} `json:"days_config"`
			}
			json.Unmarshal(rule.Recurring, &recurring)

			totalSlots := 0

			// Новый формат: days_config
			if len(recurring.DaysConfig) > 0 {
				for _, dc := range recurring.DaysConfig {
					if !dc.IsWorking || dc.StartTime == "" || dc.EndTime == "" {
						continue
					}
					startParts := strings.Split(dc.StartTime, ":")
					endParts := strings.Split(dc.EndTime, ":")
					startH, _ := strconv.Atoi(startParts[0])
					startM, _ := strconv.Atoi(startParts[1])
					endH, _ := strconv.Atoi(endParts[0])
					endM, _ := strconv.Atoi(endParts[1])
					dayMins := (endH*60 + endM) - (startH*60 + startM)
					if dayMins > 0 && dc.SlotDuration > 0 {
						totalSlots += dayMins / (dc.SlotDuration + dc.BreakBetween)
					}
				}
			} else if len(recurring.Days) > 0 && recurring.SlotDuration > 0 && recurring.StartTime != "" && recurring.EndTime != "" {
				// Старый формат: flat days
				startParts := strings.Split(recurring.StartTime, ":")
				endParts := strings.Split(recurring.EndTime, ":")

				startH, _ := strconv.Atoi(startParts[0])
				startM, _ := strconv.Atoi(startParts[1])
				endH, _ := strconv.Atoi(endParts[0])
				endM, _ := strconv.Atoi(endParts[1])

				dayMins := (endH*60 + endM) - (startH*60 + startM)
				if dayMins > 0 {
					slotsPerDay := dayMins / (recurring.SlotDuration + recurring.BreakBetween)
					totalSlots = slotsPerDay * len(recurring.Days)
				}
			}

			occupancy := 0.0
			if totalSlots > 0 {
				occupancy = float64(count) / float64(totalSlots) * 100
				if occupancy > 100 {
					occupancy = 100 // Кап на 100% на случай овербукинга
				}
			}

			analytics = append(analytics, ResourceAnalytics{
				ID:               rule.ResourceID.String(),
				Name:             item.Name,
				BookedSlotsWeek:  int(count),
				TotalSlotsWeek:   totalSlots,
				OccupancyPercent: math.Round(occupancy*100) / 100,
			})
		}

		c.JSON(http.StatusOK, gin.H{"resources": analytics})
	}
}