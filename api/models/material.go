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
	Estado       bool           `gorm:"default:false;not null" json:"estado"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relaciones
	Creador                Usuario                 `gorm:"foreignKey:CreadorID;references:GoogleID" json:"creador"`
	Colaboradores          []Usuario               `gorm:"many2many:colaboradores_material;foreignKey:ID;joinForeignKey:MaterialID;references:GoogleID;joinReferences:UsuarioID" json:"colaboradores"`
	Pasos                  []PasoMaterial          `gorm:"foreignKey:MaterialID" json:"pasos"`
	Galeria                []GaleriaMaterial       `gorm:"foreignKey:MaterialID" json:"galeria"`
	PropiedadesMecanicas   *PropiedadesMecanicas   `gorm:"foreignKey:MaterialID" json:"prop_mecanicas,omitempty"`
	PropiedadesPerceptivas *PropiedadesPerceptivas `gorm:"foreignKey:MaterialID" json:"prop_perceptivas,omitempty"`
	PropiedadesEmocionales *PropiedadesEmocionales `gorm:"foreignKey:MaterialID" json:"prop_emocionales,omitempty"`
}
