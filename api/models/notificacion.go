package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notificacion struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relación con Usuario (La clave foránea es GoogleID que es string)
	UsuarioID string `gorm:"not null" json:"usuario_id"`

	// Relación con Material (Puede ser null, por eso usamos puntero *)
	MaterialID *uuid.UUID `gorm:"type:uuid" json:"material_id"`

	Titulo  string `gorm:"not null" json:"titulo"`
	Mensaje string `gorm:"not null" json:"mensaje"`
	Leido   bool   `gorm:"default:false" json:"leido"`
	Tipo    string `json:"tipo"` // Ej: 'aprobado', 'rechazo', 'info'
}

// Aseguramos que GORM use el nombre exacto de la tabla en Supabase
func (Notificacion) TableName() string {
	return "notificaciones"
}
