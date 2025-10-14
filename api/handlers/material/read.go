package material

import (
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMaterials lista todos los materiales con relaciones
func GetMaterials(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var materials []models.Material
	if err := db.Preload("Creador").Preload("Colaboradores").Preload("Pasos").Preload("Galeria").Preload("PropiedadesMecanicas").Preload("PropiedadesPerceptivas").Preload("PropiedadesEmocionales").Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando materiales: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, materials)
}

// GetMaterial obtiene un material por ID con relaciones
func GetMaterial(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var material models.Material
	if err := db.Preload("Creador").Preload("Colaboradores").Preload("Pasos").Preload("Galeria").Preload("PropiedadesMecanicas").Preload("PropiedadesPerceptivas").Preload("PropiedadesEmocionales").First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	c.JSON(http.StatusOK, material)
}

// Estructura resumida para la respuesta
type SummaryMaterial struct {
	ID                   uuid.UUID          `json:"id"`
	Nombre               string             `json:"nombre"`
	Descripcion          string             `json:"descripcion"`
	Composicion          models.StringArray `json:"composicion"`
	DerivadoDe           uuid.UUID          `json:"derivado_de"`
	PrimeraImagenGaleria string             `json:"primera_imagen_galeria,omitempty"`
}

// GetMaterialsSummary lista los materiales con información resumida
func GetMaterialsSummary(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var materials []models.Material
	if err := db.Preload("Galeria").Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando materiales resumidos: " + err.Error()})
		return
	}

	var summaries []SummaryMaterial
	for _, m := range materials {
		primeraImagen := ""
		if len(m.Galeria) > 0 {
			primeraImagen = m.Galeria[0].URLImagen
		}

		summaries = append(summaries, SummaryMaterial{
			ID:                   m.ID,
			Nombre:               m.Nombre,
			Descripcion:          m.Descripcion,
			Composicion:          m.Composicion,
			DerivadoDe:           m.DerivadoDe,
			PrimeraImagenGaleria: primeraImagen,
		})
	}

	c.JSON(http.StatusOK, summaries)
}
