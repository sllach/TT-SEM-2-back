package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GaleriaMaterial struct {
	gorm.Model
	MaterialID uuid.UUID `gorm:"type:uuid;not null" json:"material_id"`
	URLImagen  string    `gorm:"size:512;not null" json:"url_imagen"`
	Caption    string    `gorm:"type:text" json:"caption"`
}
