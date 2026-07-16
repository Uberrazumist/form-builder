package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/datatypes"
)

// User – модель пользователя
type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Email     string    `gorm:"unique;not null" json:"email"`
    Password  string    `gorm:"not null" json:"-"`
    FullName  string    `json:"full_name"`
    Role      string    `gorm:"default:'viewer'" json:"role"`
    Verified  bool      `gorm:"default:false" json:"verified"`
    CreatedAt time.Time `json:"created_at"`
}

// EmailVerification – подтверждение email и сброс пароля
type EmailVerification struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Email     string    `gorm:"not null;index" json:"email"`
    Code      string    `gorm:"not null" json:"code"`
    Type      string    `gorm:"not null" json:"type"` // 'registration' или 'reset_password'
    ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
    Used      bool      `gorm:"default:false" json:"used"`
    CreatedAt time.Time `json:"created_at"`
}

// Form – модель формы
type Form struct {
    ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Title       string         `gorm:"not null" json:"title"`
    Description string         `json:"description"`
    CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    IsPublished bool           `gorm:"default:false" json:"is_published"`
    IsPublic    bool           `gorm:"default:false" json:"is_public"`
    Settings    datatypes.JSON `gorm:"type:jsonb" json:"settings,omitempty"`
    Questions   []Question     `gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE" json:"questions"`
}

// Question – модель вопроса (исправлено: убраны omitempty для DependsOn и DictionaryID)
type Question struct {
    ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    FormID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"form_id"`
    Type          string         `gorm:"not null" json:"type"`
    Title         string         `gorm:"not null" json:"title"`
    Description   string         `json:"description"`
    OrderIndex    int            `gorm:"not null" json:"order_index"`
    IsRequired    bool           `gorm:"default:false" json:"is_required"`
    Options       datatypes.JSON `gorm:"type:jsonb" json:"options,omitempty"`
    Validation    datatypes.JSON `gorm:"type:jsonb" json:"validation,omitempty"`
    DependsOn     *uuid.UUID     `gorm:"type:uuid;index" json:"depends_on"`          // omitempty удалён
    DependsValues datatypes.JSON `gorm:"type:jsonb" json:"depends_values,omitempty"`
    DictionaryID  *uuid.UUID     `gorm:"type:uuid;index" json:"dictionary_id"`       // omitempty удалён
    IsBooking     bool           `gorm:"default:false" json:"is_booking"`
}

// Response – ответ пользователя
type Response struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    FormID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"form_id"`
    UserID    *uuid.UUID     `gorm:"type:uuid" json:"user_id,omitempty"`
    Answers   datatypes.JSON `gorm:"type:jsonb" json:"answers"`
    CreatedAt time.Time      `json:"created_at"`
}

// Dictionary – справочник
type Dictionary struct {
    ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Name        string         `gorm:"not null" json:"name"`
    Description string         `json:"description"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    Items       []DictionaryItem `gorm:"foreignKey:DictionaryID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

// DictionaryItem – элемент справочника (исправлено: убран omitempty для ParentID)
type DictionaryItem struct {
    ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    DictionaryID uuid.UUID      `gorm:"type:uuid;not null;index" json:"dictionary_id"`
    ParentID     *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id"`            // omitempty удалён
    Name         string         `gorm:"not null" json:"name"`
    Code         string         `json:"code"`
    Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
}

// Booking – бронирование (новая структура с уникальным индексом)
type Booking struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    FormID     uuid.UUID `gorm:"type:uuid;not null;index" json:"form_id"`
    UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
    ResourceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_booking_resource_time" json:"resource_id"`
    Date       time.Time `gorm:"type:date;not null;uniqueIndex:idx_booking_resource_time" json:"date"`
    StartTime  time.Time `gorm:"not null;uniqueIndex:idx_booking_resource_time" json:"start_time"`
    EndTime    time.Time `gorm:"not null" json:"end_time"`
    CreatedAt  time.Time `json:"created_at"`
}

// Class – класс (для обратной совместимости)
type Class struct {
    ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Name string    `gorm:"unique;not null" json:"name"`
}

// Teacher – учитель (для обратной совместимости)
type Teacher struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    FullName  string         `gorm:"not null" json:"full_name"`
    ClassIDs  datatypes.JSON `gorm:"type:jsonb" json:"class_ids,omitempty"`
    Available bool           `gorm:"default:true" json:"available"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}

// TimeSlot – временной слот (для обратной совместимости)
type TimeSlot struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    TeacherID uuid.UUID `gorm:"type:uuid;index" json:"teacher_id"`
    Date      time.Time `gorm:"not null" json:"date"`
    StartTime time.Time `gorm:"not null" json:"start_time"`
    EndTime   time.Time `gorm:"not null" json:"end_time"`
    IsBooked  bool      `gorm:"default:false" json:"is_booked"`
}

// FormPermission – права доступа к форме
type FormPermission struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    FormID     uuid.UUID `gorm:"type:uuid;not null;index" json:"form_id"`
    UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
    Permission string    `gorm:"not null" json:"permission"` // "view", "edit"
    CreatedAt  time.Time `json:"created_at"`
}

// ScheduleRule – правило расписания для генерации слотов на лету
type ScheduleRule struct {
    ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    ResourceID uuid.UUID      `gorm:"type:uuid;not null;index" json:"resource_id"` // ID элемента справочника (учитель, кабинет)
    Name       string         `gorm:"not null" json:"name"`                         // "Расписание Иванова"
    Recurring  datatypes.JSON `gorm:"type:jsonb;not null" json:"recurring"`         // Правила генерации слотов
    IsDeleted  bool           `gorm:"default:false" json:"is_deleted"`              // Мягкое удаление
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
}
