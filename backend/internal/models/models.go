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
    Settings    datatypes.JSON `gorm:"type:jsonb"` // <-- исправлено
}

type Question struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID      uuid.UUID `gorm:"type:uuid;not null;index"`
    Type        string    `gorm:"not null"`
    Title       string    `gorm:"not null"`
    Description string
    OrderIndex  int       `gorm:"not null"`
    IsRequired  bool      `gorm:"default:false"`
    Options     datatypes.JSON `gorm:"type:jsonb"`
    Validation  datatypes.JSON `gorm:"type:jsonb"`
    DependsOn   *uuid.UUID `gorm:"type:uuid;index"`
    DependsValues datatypes.JSON `gorm:"type:jsonb"`
}

type Response struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID    uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID    *uuid.UUID `gorm:"type:uuid"`
    Answers   datatypes.JSON `gorm:"type:jsonb"`
    CreatedAt time.Time
    IPAddress string
}

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

type Booking struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TimeSlotID uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID     uuid.UUID `gorm:"type:uuid;not null"`
    FormID     uuid.UUID `gorm:"type:uuid;not null"`
    CreatedAt  time.Time
}

type FormPermission struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    FormID     uuid.UUID `gorm:"type:uuid;not null;index"`
    UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
    Permission string    `gorm:"not null"`
}
