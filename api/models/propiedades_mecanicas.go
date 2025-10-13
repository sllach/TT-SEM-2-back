package models

import "github.com/google/uuid"

type PropiedadesMecanicas struct {
	MaterialID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	Resistencia string    `gorm:"type:text" json:"resistencia"`
	Dureza      string    `gorm:"type:text" json:"dureza"`
	Elasticidad string    `gorm:"type:text" json:"elasticidad"`
	Ductilidad  string    `gorm:"type:text" json:"ductilidad"`
	Fragilidad  string    `gorm:"type:text" json:"fragilidad"`
}
