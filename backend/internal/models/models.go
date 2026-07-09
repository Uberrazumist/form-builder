package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
)

type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email     string    `gorm:"unique;not null"`
    Password  string    `gorm:"not null"`
    FullName  string
    Role      string    `gorm:"default:'viewer'"`
    Verified  bool      `gorm:"default:false"`
    CreatedAt time.Time
}

type EmailVerification struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email     string    `gorm:"not null;index"`
    Code      string    `gorm:"not null"`
    Type      string    `gorm:"not null"` // 'registration' или 'reset_password'
    ExpiresAt time.Time `gorm:"not null"`
    Used      bool      `gorm:"default:false"`
    CreatedAt time.Time
}

type Form struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Title       string    `gorm:"not null"`
    Description string
    CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    IsPublished bool      `gorm:"default:false"`
    IsPublic    bool      `gorm:"default:false"`
    Settings    datatypes.JSON `gorm:"type:jsonb"`
    Questions   []Question `gorm:"foreignKey:FormID"`
}

type Question struct {
    ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID          uuid.UUID `gorm:"type:uuid;not null;index"`
    Type            string    `gorm:"not null"` // text, textarea, radio, checkbox, select, rating, dictionary
    Title           string    `gorm:"not null"`
    Description     string
    OrderIndex      int       `gorm:"not null"`
    IsRequired      bool      `gorm:"default:false"`
    Options         datatypes.JSON `gorm:"type:jsonb"` // для radio/checkbox/select
    Validation      datatypes.JSON `gorm:"type:jsonb"`
    DependsOn       *uuid.UUID `gorm:"type:uuid;index"`
    DependsValues   datatypes.JSON `gorm:"type:jsonb"`
    // Новые поля для справочников
    DictionaryID    *uuid.UUID `gorm:"type:uuid;index"` // ссылка на справочник
    FilterMetadata  datatypes.JSON `gorm:"type:jsonb"`   // правила фильтрации, например {"parent_question_id": "uuid"}
    IsBooking       bool       `gorm:"default:false"`    // если true, то при сохранении ответа проверяем занятость
}

type Response struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID    uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID    *uuid.UUID `gorm:"type:uuid"`
    Answers   datatypes.JSON `gorm:"type:jsonb"`
    CreatedAt time.Time
}

// --- Справочники ---
type Dictionary struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name        string    `gorm:"not null"`
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Items       []DictionaryItem `gorm:"foreignKey:DictionaryID"`
}

type DictionaryItem struct {
    ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    DictionaryID uuid.UUID `gorm:"type:uuid;not null;index"`
    ParentID     *uuid.UUID `gorm:"type:uuid;index"` // для иерархии
    Name         string    `gorm:"not null"`
    Code         string    // уникальный код внутри справочника (опционально)
    Metadata     datatypes.JSON `gorm:"type:jsonb"` // любые доп. поля (class_id, teacher_id, time_start, etc.)
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type Booking struct {
    ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    DictionaryItemID uuid.UUID `gorm:"type:uuid;not null;index"`
    FormID           uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID           uuid.UUID `gorm:"type:uuid;not null"`
    CreatedAt        time.Time
}

// Старые модели (Class, Teacher, TimeSlot) пока оставляем, но в будущем заменим на Dictionary
type Class struct {
    ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name string    `gorm:"unique;not null"`
}

type Teacher struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FullName  string    `gorm:"not null"`
    ClassID   uuid.UUID `gorm:"type:uuid;index"`
    Available bool      `gorm:"default:true"`
}

type TimeSlot struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TeacherID uuid.UUID `gorm:"type:uuid;not null;index"`
    Date      time.Time `gorm:"not null"`
    StartTime time.Time `gorm:"not null"`
    EndTime   time.Time `gorm:"not null"`
    IsBooked  bool      `gorm:"default:false"`
}

type FormPermission struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID     uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
    Permission string    `gorm:"not null"`
}
