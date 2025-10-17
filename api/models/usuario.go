package models

import (
	"time"

	"gorm.io/gorm"
)

type Usuario struct {
	GoogleID string `gorm:"primaryKey;unique;not null" json:"google_id"`
	Nombre   string `gorm:"not null" json:"nombre"`
	Email    string `gorm:"unique;not null" json:"email"`
	Rol      string `gorm:"default:'Lector'" json:"rol"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
