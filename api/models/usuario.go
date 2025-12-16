package models

import (
	"time"

	"gorm.io/gorm"
)

type Usuario struct {
	// Quitamos gorm.Model para evitar que cree un ID uint automatico

	GoogleID   string `gorm:"primaryKey;type:text" json:"google_id"` // <--- ESTO ES LA CLAVE AHORA
	SupabaseID string `gorm:"type:text;unique" json:"supabase_id"`
	Nombre     string `gorm:"size:255;not null" json:"nombre"`
	Email      string `gorm:"size:255;not null;unique" json:"email"`
	Rol        string `gorm:"default:'lector'" json:"rol"` // lector, colaborador, administrador

	// Agregamos manualmente los campos de tiempo que quitamos al sacar gorm.Model
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
