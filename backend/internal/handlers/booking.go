package handlers

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

func GetAvailableSlots(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        teacherIDStr := c.Query("teacher_id")
        dateStr := c.Query("date")

        if teacherIDStr == "" || dateStr == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "teacher_id and date are required"})
            return
        }

        teacherID, err := uuid.Parse(teacherIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid teacher_id"})
            return
        }

        date, err := time.Parse("2006-01-02", dateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
            return
        }

        // Находим справочник "Время"
        var timeDict models.Dictionary
        if err := db.Where("name = ?", "Время").First(&timeDict).Error; err != nil {
            if err := db.Where("name = ?", "Time").First(&timeDict).Error; err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "Time dictionary not found"})
                return
            }
        }

        var allSlots []models.DictionaryItem
        if err := db.Where("dictionary_id = ?", timeDict.ID).Order("name").Find(&allSlots).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch time slots"})
            return
        }

        // Получаем занятые слоты
        var bookedSlotIDs []uuid.UUID
        if err := db.Model(&models.Booking{}).
            Where("teacher_id = ? AND date = ?", teacherID, date).
            Pluck("slot_id", &bookedSlotIDs).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
            return
        }

        bookedMap := make(map[uuid.UUID]bool)
        for _, id := range bookedSlotIDs {
            bookedMap[id] = true
        }

        availableSlots := []models.DictionaryItem{}
        for _, slot := range allSlots {
            if !bookedMap[slot.ID] {
                availableSlots = append(availableSlots, slot)
            }
        }

        c.JSON(http.StatusOK, availableSlots)
    }
}
