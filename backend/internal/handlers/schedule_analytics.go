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

		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())
		sunday := monday.AddDate(0, 0, 6)
		sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, now.Location())

		for _, rule := range rules {
			var config models.RecurringSchedule
			if err := json.Unmarshal(rule.Recurring, &config); err != nil {
				// Пропускаем правило с невалидным JSON
				continue
			}

			var item models.DictionaryItem
			db.First(&item, "id = ?", rule.ResourceID)

			// Считаем слоты по новым weekly_intervals
			totalSlots := 0
			for _, day := range config.WeeklyIntervals {
				for _, interval := range day.Intervals {
					startParts := strings.Split(interval.Start, ":")
					endParts := strings.Split(interval.End, ":")

					sh, _ := strconv.Atoi(startParts[0])
					sm, _ := strconv.Atoi(startParts[1])
					eh, _ := strconv.Atoi(endParts[0])
					em, _ := strconv.Atoi(endParts[1])

					dayMins := (eh*60 + em) - (sh*60 + sm)
					if dayMins > 0 && config.SlotDuration > 0 {
						totalSlots += dayMins / (config.SlotDuration + config.BreakBetween)
					}
				}
			}

			// Если weekly_intervals пустой, пробуем старый формат days_config
			if totalSlots == 0 && len(config.WeeklyIntervals) == 0 {
				// Попытка прочитать старый формат
				var oldFormat struct {
					Days         []int  `json:"days"`
					SlotDuration int    `json:"slot_duration"`
					BreakBetween int    `json:"break_between"`
					StartTime    string `json:"start_time"`
					EndTime      string `json:"end_time"`
				}
				if err := json.Unmarshal(rule.Recurring, &oldFormat); err == nil {
					if len(oldFormat.Days) > 0 && oldFormat.SlotDuration > 0 && oldFormat.StartTime != "" && oldFormat.EndTime != "" {
						startParts := strings.Split(oldFormat.StartTime, ":")
						endParts := strings.Split(oldFormat.EndTime, ":")
						sh, _ := strconv.Atoi(startParts[0])
						sm, _ := strconv.Atoi(startParts[1])
						eh, _ := strconv.Atoi(endParts[0])
						em, _ := strconv.Atoi(endParts[1])
						dayMins := (eh*60 + em) - (sh*60 + sm)
						if dayMins > 0 {
							slotsPerDay := dayMins / (oldFormat.SlotDuration + oldFormat.BreakBetween)
							totalSlots = slotsPerDay * len(oldFormat.Days)
						}
					}
				}
			}

			var count int64
			db.Model(&models.Booking{}).
				Where("resource_id = ? AND date >= ? AND date <= ?", rule.ResourceID, monday, sunday).
				Count(&count)

			occupancy := 0.0
			if totalSlots > 0 {
				occupancy = float64(count) / float64(totalSlots) * 100
				if occupancy > 100 {
					occupancy = 100
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
