package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- TIPOS JSON (Dinámicos) ---

// 1. Composición
type Componente struct {
	Elemento string `json:"elemento"`
	Cantidad string `json:"cantidad"`
}
type JSONComponentes []Componente

func (j JSONComponentes) Value() (driver.Value, error) { return json.Marshal(j) }
func (j *JSONComponentes) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}

// 2. Mecánicas
type PropiedadMecanica struct {
	Nombre string `json:"nombre"`
	Valor  string `json:"valor"`
	Unidad string `json:"unidad"`
}
type JSONMecanicas []PropiedadMecanica

func (j JSONMecanicas) Value() (driver.Value, error) { return json.Marshal(j) }
func (j *JSONMecanicas) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}

// 3. Generales (Perceptivas y Emocionales usan la misma estructura)
type PropiedadGeneral struct {
	Nombre string `json:"nombre"`
	Valor  string `json:"valor"`
}
type JSONGenerales []PropiedadGeneral

func (j JSONGenerales) Value() (driver.Value, error) { return json.Marshal(j) }
func (j *JSONGenerales) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}

// --- MODELO MATERIAL ---
type Material struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Nombre      string    `gorm:"size:255;not null" json:"nombre"`
	Descripcion string    `gorm:"type:text" json:"descripcion"`

	// Columnas JSONB
	Composicion            JSONComponentes `gorm:"type:jsonb" json:"composicion"`
	PropiedadesMecanicas   JSONMecanicas   `gorm:"type:jsonb" json:"prop_mecanicas"`
	PropiedadesPerceptivas JSONGenerales   `gorm:"type:jsonb" json:"prop_perceptivas"`
	PropiedadesEmocionales JSONGenerales   `gorm:"type:jsonb" json:"prop_emocionales"`

	Herramientas StringArray `gorm:"type:jsonb" json:"herramientas"`

	// Relaciones
	CreadorID string  `gorm:"type:text;not null" json:"creador_id"`
	Creador   Usuario `gorm:"foreignKey:CreadorID;references:GoogleID" json:"creador"`

	DerivadoDe uuid.UUID `gorm:"type:uuid;default:null" json:"derivado_de"`
	Estado     bool      `gorm:"default:false" json:"estado"`

	Colaboradores []Usuario `gorm:"many2many:material_colaboradores;joinForeignKey:MaterialID;joinReferences:UsuarioID;references:GoogleID" json:"colaboradores"`

	Pasos   []PasoMaterial    `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE" json:"pasos"`
	Galeria []GaleriaMaterial `gorm:"foreignKey:MaterialID;constraint:OnDelete:CASCADE" json:"galeria"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

