package models

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type PropiedadesPerceptivas struct {
	MaterialID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	Color            string    `gorm:"type:text" json:"color"`
	Brillo           string    `gorm:"type:text" json:"brillo"`
	Textura          string    `gorm:"type:text" json:"textura"`
	Transparencia    string    `gorm:"type:text" json:"transparencia"`
	SensacionTermica string    `gorm:"type:text" json:"sensacion_termica"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
