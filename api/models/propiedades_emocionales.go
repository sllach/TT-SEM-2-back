package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PropiedadesEmocionales struct {
	MaterialID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	CalidezEmocional        string    `gorm:"type:text" json:"calidez_emocional"`
	Inspiracion             string    `gorm:"type:text" json:"inspiracion"`
	SostenibilidadPercibida string    `gorm:"type:text" json:"sostenibilidad_percibida"`
	Armonia                 string    `gorm:"type:text" json:"armonia"`
	InnovacionEmocional     string    `gorm:"type:text" json:"innovacion_emocional"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
