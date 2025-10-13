package models

import "gorm.io/gorm"

type Usuario struct {
	gorm.Model
	GoogleID string `gorm:"unique;not null" json:"google_id"`
	Nombre   string `gorm:"not null" json:"nombre"`
	Email    string `gorm:"unique;not null" json:"email"`
}
