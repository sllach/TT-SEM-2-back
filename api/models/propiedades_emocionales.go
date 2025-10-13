package models

import "github.com/google/uuid"

type PropiedadesEmocionales struct {
	MaterialID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	CalidezEmocional        string    `gorm:"type:text" json:"calidez_emocional"`
	Inspiracion             string    `gorm:"type:text" json:"inspiracion"`
	SostenibilidadPercibida string    `gorm:"type:text" json:"sostenibilidad_percibida"`
	Armonia                 string    `gorm:"type:text" json:"armonia"`
	InnovacionEmocional     string    `gorm:"type:text" json:"innovacion_emocional"`
}
