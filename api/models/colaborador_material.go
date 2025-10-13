package models

import "github.com/google/uuid"

type ColaboradorMaterial struct {
	MaterialID uuid.UUID `gorm:"type:uuid;primaryKey" json:"material_id"`
	UsuarioID  uint      `gorm:"primaryKey" json:"usuario_id"`
}
