package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ColaboradorMaterial struct {
	MaterialID uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	UsuarioID  string    `gorm:"primaryKey" json:"usuario_id"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
