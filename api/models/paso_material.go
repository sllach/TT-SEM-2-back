package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasoMaterial struct {
	gorm.Model
	MaterialID  uuid.UUID `gorm:"type:uuid;not null" json:"material_id"`
	OrdenPaso   int       `gorm:"not null" json:"orden_paso"`
	Descripcion string    `gorm:"type:text;not null" json:"descripcion"`
	URLImagen   string    `gorm:"size:512" json:"url_imagen"`
	URLVideo    string    `gorm:"size:512" json:"url_video"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
