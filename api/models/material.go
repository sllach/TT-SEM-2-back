package models

import (
	"time"

	"github.com/google/uuid"
)

type Material struct {
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Nombre        string      `gorm:"not null" json:"nombre"`
	Descripcion   string      `gorm:"type:text" json:"descripcion"`
	Herramientas  StringArray `gorm:"type:jsonb;default:'[]'::jsonb" json:"herramientas"`
	Composicion   StringArray `gorm:"type:jsonb;default:'[]'::jsonb" json:"composicion"`
	DerivadoDe    uuid.UUID   `gorm:"type:uuid" json:"derivado_de"`
	CreadorID     uint        `gorm:"not null" json:"creador_id"`
	CreadoEn      time.Time   `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn time.Time   `gorm:"autoUpdateTime" json:"actualizado_en"`

	// Relaciones
	Creador                Usuario                 `gorm:"foreignKey:CreadorID"`
	Colaboradores          []Usuario               `gorm:"many2many:colaboradores_material;"`
	Pasos                  []PasoMaterial          `gorm:"foreignKey:MaterialID"`
	Galeria                []GaleriaMaterial       `gorm:"foreignKey:MaterialID"`
	PropiedadesMecanicas   *PropiedadesMecanicas   `gorm:"foreignKey:MaterialID"`
	PropiedadesPerceptivas *PropiedadesPerceptivas `gorm:"foreignKey:MaterialID"`
	PropiedadesEmocionales *PropiedadesEmocionales `gorm:"foreignKey:MaterialID"`
}
