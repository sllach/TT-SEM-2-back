package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Definición explícita de la tabla intermedia
type ColaboradorMaterial struct {
	MaterialID uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	UsuarioID  string    `gorm:"type:text;primaryKey" json:"usuario_id"` // TEXTO, NO BIGINT

	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Forzamos el nombre exacto de la tabla
func (ColaboradorMaterial) TableName() string {
	return "material_colaboradores"
}
