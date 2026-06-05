package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет пользователя системы
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string         `gorm:"unique;not null" json:"email" example:"user@example.com"`
	Password  string         `gorm:"not null" json:"-" swaggerignore:"true"` // Скрыто в Swagger
	Salt      string         `gorm:"not null" json:"-" swaggerignore:"true"` // Скрыто в Swagger
	YandexID  string         `gorm:"uniqueIndex" json:"-" swaggerignore:"true"`
	VKID      string         `json:"-" swaggerignore:"true"`
	CreatedAt time.Time      `json:"created_at" example:"2024-01-15T12:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2024-01-15T12:00:00Z"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
