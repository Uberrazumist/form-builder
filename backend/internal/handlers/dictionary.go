package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

// --- Справочники ---
func ListDictionaries(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var dicts []models.Dictionary
        if err := db.Find(&dicts).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dictionaries"})
            return
        }
        c.JSON(http.StatusOK, dicts)
    }
}

func CreateDictionary(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            Name        string `json:"name" binding:"required"`
            Description string `json:"description"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        dict := models.Dictionary{
            Name:        input.Name,
            Description: input.Description,
        }
        if err := db.Create(&dict).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dictionary"})
            return
        }
        c.JSON(http.StatusCreated, dict)
    }
}

func GetDictionary(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var dict models.Dictionary
        if err := db.Preload("Items").First(&dict, "id = ?", id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Dictionary not found"})
            return
        }
        c.JSON(http.StatusOK, dict)
    }
}

func UpdateDictionary(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var dict models.Dictionary
        if err := db.First(&dict, "id = ?", id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Dictionary not found"})
            return
        }
        var input struct {
            Name        string `json:"name"`
            Description string `json:"description"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        if input.Name != "" {
            dict.Name = input.Name
        }
        if input.Description != "" {
            dict.Description = input.Description
        }
        db.Save(&dict)
        c.JSON(http.StatusOK, dict)
    }
}

func DeleteDictionary(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        if err := db.Where("dictionary_id = ?", id).Delete(&models.DictionaryItem{}).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete items"})
            return
        }
        if err := db.Delete(&models.Dictionary{}, "id = ?", id).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dictionary"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Dictionary deleted"})
    }
}

// --- Элементы справочника ---
func ListDictionaryItems(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        dictID := c.Param("id")
        parentID := c.Query("parent")
        filterMetadata := c.Query("filter_metadata")

        query := db.Where("dictionary_id = ?", dictID)
        if parentID != "" {
            query = query.Where("parent_id = ?", parentID)
        }
        if filterMetadata != "" {
            var meta map[string]interface{}
            if err := json.Unmarshal([]byte(filterMetadata), &meta); err == nil {
                for k, v := range meta {
                    // Используем оператор @> для проверки вхождения ключа
                    valStr, ok := v.(string)
                    if !ok {
                        continue
                    }
                    query = query.Where("metadata @> ?", datatypes.JSON([]byte(`{"`+k+`":"`+valStr+`"}`)))
                }
            }
        }

        var items []models.DictionaryItem
        if err := query.Find(&items).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
            return
        }
        c.JSON(http.StatusOK, items)
    }
}

func CreateDictionaryItem(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        dictID := c.Param("id")
        var input struct {
            ParentID *string                 `json:"parent_id"`
            Name     string                  `json:"name" binding:"required"`
            Code     string                  `json:"code"`
            Metadata map[string]interface{}  `json:"metadata"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        item := models.DictionaryItem{
            DictionaryID: uuid.MustParse(dictID),
            Name:         input.Name,
            Code:         input.Code,
        }
        if input.ParentID != nil {
            pid := uuid.MustParse(*input.ParentID)
            item.ParentID = &pid
        }
        if input.Metadata != nil {
            jsonData, err := json.Marshal(input.Metadata)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize metadata"})
                return
            }
            item.Metadata = datatypes.JSON(jsonData)
        }
        if err := db.Create(&item).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
            return
        }
        c.JSON(http.StatusCreated, item)
    }
}

func DeleteDictionaryItem(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        itemID := c.Param("itemId")
        if err := db.Delete(&models.DictionaryItem{}, "id = ?", itemID).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
    }
}

// --- Проверка занятости ---
func CheckBooking(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        itemID := c.Query("item_id")
        formID := c.Query("form_id")
        if itemID == "" || formID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "item_id and form_id required"})
            return
        }
        var count int64
        if err := db.Model(&models.Booking{}).Where("dictionary_item_id = ? AND form_id = ?", itemID, formID).Count(&count).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check booking"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"booked": count > 0})
    }
}
