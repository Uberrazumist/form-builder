package handlers

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/datatypes"
    "gorm.io/gorm"

    "github.com/Uberrazumist/form-builder/backend/internal/models"
)

// --- CRUD для справочников ---
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

// UpdateDictionary – только авторизованным (любой может редактировать)
func UpdateDictionary(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var dict models.Dictionary
		if err := db.First(&dict, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Dictionary not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
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
		if err := db.Save(&dict).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dictionary"})
			return
		}
		c.JSON(http.StatusOK, dict)
	}
}

func DeleteDictionary(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
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
            // Парсим JSON фильтра
            var filters map[string]interface{}
            if err := json.Unmarshal([]byte(filterMetadata), &filters); err == nil {
                // Ищем элементы, у которых metadata содержит все ключи-значения из фильтра
                // Для PostgreSQL используем оператор @> (contains)
                for key, val := range filters {
                    // Безопасная конвертация значения в строку
                    var valStr string
                    switch v := val.(type) {
                    case string:
                        valStr = v
                    case float64: // JSON числа парсятся как float64
                        valStr = fmt.Sprintf("%v", v)
                    case bool:
                        valStr = fmt.Sprintf("%v", v)
                    default:
                        continue // Пропускаем неподдерживаемые типы
                    }
                    query = query.Where("metadata @> ?", datatypes.JSON([]byte(`{"`+key+`":"`+valStr+`"}`)))
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
            ParentID *string                `json:"parent_id"`
            Name     string                 `json:"name" binding:"required"`
            Code     string                 `json:"code"`
            Metadata map[string]interface{} `json:"metadata"`
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
        if input.ParentID != nil && *input.ParentID != "" {
            pid, err := uuid.Parse(*input.ParentID)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат parent_id"})
                return
            }
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

		// Сначала удаляем все дочерние элементы (каскадное удаление)
		if err := db.Where("parent_id = ?", itemID).Delete(&models.DictionaryItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete child items"})
			return
		}

		// Затем удаляем сам элемент
		if err := db.Delete(&models.DictionaryItem{}, "id = ?", itemID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
	}
}

func UpdateDictionaryItem(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        itemID := c.Param("itemId")
        var input struct {
            Name     string                 `json:"name"`
            Code     string                 `json:"code"`
            ParentID *string                `json:"parent_id"`
            Metadata map[string]interface{} `json:"metadata"`
        }
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var item models.DictionaryItem
        if err := db.First(&item, "id = ?", itemID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
            return
        }

        if input.Name != "" {
            item.Name = input.Name
        }
        if input.Code != "" {
            item.Code = input.Code
        }
        if input.ParentID != nil {
            if *input.ParentID == "" {
                item.ParentID = nil
            } else {
                pid, err := uuid.Parse(*input.ParentID)
                if err != nil {
                    c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат parent_id"})
                    return
                }
                item.ParentID = &pid
            }
        } else {
            // Явно сбрасываем parent_id в nil, если фронтенд прислал parent_id: null
            item.ParentID = nil
        }
        if input.Metadata != nil {
            jsonData, err := json.Marshal(input.Metadata)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize metadata"})
                return
            }
            item.Metadata = datatypes.JSON(jsonData)
        }

        if err := db.Save(&item).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
            return
        }

        c.JSON(http.StatusOK, item)
    }
}
