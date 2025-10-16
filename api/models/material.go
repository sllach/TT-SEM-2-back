package models

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type Material struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nombre       string         `gorm:"not null" json:"nombre"`
	Descripcion  string         `gorm:"type:text" json:"descripcion"`
	Herramientas StringArray    `gorm:"type:jsonb;default:'[]'::jsonb" json:"herramientas"`
	Composicion  StringArray    `gorm:"type:jsonb;default:'[]'::jsonb" json:"composicion"`
	DerivadoDe   uuid.UUID      `gorm:"type:uuid" json:"derivado_de"`
	CreadorID    string         `gorm:"not null" json:"creador_id"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Creador                Usuario                 `gorm:"foreignKey:CreadorID;references:GoogleID"`                                                                              // Referencia a GoogleID
	Colaboradores          []Usuario               `gorm:"many2many:colaboradores_material;foreignKey:ID;joinForeignKey:MaterialID;references:GoogleID;joinReferences:UsuarioID"` // Corregido
	Pasos                  []PasoMaterial          `gorm:"foreignKey:MaterialID"`
	Galeria                []GaleriaMaterial       `gorm:"foreignKey:MaterialID"`
	PropiedadesMecanicas   *PropiedadesMecanicas   `gorm:"foreignKey:MaterialID"`
	PropiedadesPerceptivas *PropiedadesPerceptivas `gorm:"foreignKey:MaterialID"`
	PropiedadesEmocionales *PropiedadesEmocionales `gorm:"foreignKey:MaterialID"`
}
