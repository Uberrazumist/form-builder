package models

import (
    "time"
    "github.com/google/uuid"
)

// --- Пользователи (уже есть) ---
type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email     string    `gorm:"unique;not null"`
    Password  string    `gorm:"not null"`
    FullName  string
    Role      string    `gorm:"default:'viewer'"`
    CreatedAt time.Time
}

// --- Формы (опросы) ---
type Form struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Title       string    `gorm:"not null"`
    Description string
    CreatedBy   uuid.UUID `gorm:"type:uuid;not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    IsPublished bool      `gorm:"default:false"`
    IsPublic    bool      `gorm:"default:false"` // доступна по ссылке без авторизации
    Settings    map[string]interface{} `gorm:"type:jsonb"`
}

// --- Вопросы ---
type Question struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID      uuid.UUID `gorm:"type:uuid;not null;index"`
    Type        string    `gorm:"not null"` // text, textarea, radio, checkbox, select, rating, date, time, class_choice, teacher_choice, time_choice
    Title       string    `gorm:"not null"`
    Description string
    OrderIndex  int       `gorm:"not null"`
    IsRequired  bool      `gorm:"default:false"`
    Options     []string  `gorm:"type:jsonb"` // для radio/checkbox/select
    Validation  map[string]interface{} `gorm:"type:jsonb"` // min/max, regex, etc.
    DependsOn   *uuid.UUID `gorm:"type:uuid;index"` // ID вопроса, от которого зависит
    DependsValues []string `gorm:"type:jsonb"` // значения, при которых этот вопрос показывается
}

// --- Ответы на формы ---
type Response struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID    uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID    *uuid.UUID `gorm:"type:uuid"` // NULL для анонимных
    Answers   map[string]interface{} `gorm:"type:jsonb"` // question_id -> answer
    CreatedAt time.Time
}

// --- Классы (справочник) ---
type Class struct {
    ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Name string    `gorm:"unique;not null"` // "9А", "10Б", etc.
}

// --- Учителя (справочник) ---
type Teacher struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FullName  string    `gorm:"not null"`
    ClassID   uuid.UUID `gorm:"type:uuid;index"` // учитель закреплён за классом
    Available bool      `gorm:"default:true"`
}

// --- Временные слоты учителя ---
type TimeSlot struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TeacherID uuid.UUID `gorm:"type:uuid;not null;index"`
    Date      time.Time `gorm:"not null"` // дата слота
    StartTime time.Time `gorm:"not null"` // время начала
    EndTime   time.Time `gorm:"not null"` // время окончания
    IsBooked  bool      `gorm:"default:false"`
}

// --- Бронирование (занятый слот) ---
type Booking struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TimeSlotID uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID     uuid.UUID `gorm:"type:uuid;not null"` // кто забронировал
    FormID     uuid.UUID `gorm:"type:uuid;not null"` // из какой формы
    CreatedAt  time.Time
}

// --- Права на форму (уже было) ---
type FormPermission struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID     uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
    Permission string    `gorm:"not null"` // view, edit, admin
}
