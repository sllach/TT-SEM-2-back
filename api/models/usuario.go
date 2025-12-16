package models

import (
	"time"

	"gorm.io/gorm"
)

type Usuario struct {

	GoogleID   string `gorm:"primaryKey;type:text" json:"google_id"`
	SupabaseID string `gorm:"type:text;unique" json:"supabase_id"`
	Nombre     string `gorm:"size:255;not null" json:"nombre"`
	Email      string `gorm:"size:255;not null;unique" json:"email"`
	Rol        string `gorm:"default:'lector'" json:"rol"` // lector, colaborador, administrador

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

