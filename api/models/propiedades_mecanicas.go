package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PropiedadesMecanicas struct {
	MaterialID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	Resistencia string    `gorm:"type:text" json:"resistencia"`
	Dureza      string    `gorm:"type:text" json:"dureza"`
	Elasticidad string    `gorm:"type:text" json:"elasticidad"`
	Ductilidad  string    `gorm:"type:text" json:"ductilidad"`
	Fragilidad  string    `gorm:"type:text" json:"fragilidad"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
