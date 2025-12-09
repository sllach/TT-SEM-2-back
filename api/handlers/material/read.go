package material

import (
	"log"
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMaterials lista SOLO materiales aprobados
func GetMaterials(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var materials []models.Material
	if err := db.Where("estado = ?", true).
		Preload("Creador").
		Preload("Colaboradores").
		Preload("Pasos").
		Preload("Galeria").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando materiales: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, materials)
}

// GetMaterial obtiene un material por ID SOLO si está aprobado
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
	if err := db.Where("id = ? AND estado = ?", id, true).
		Preload("Creador").
		Preload("Colaboradores").
		Preload("Pasos").
		Preload("Galeria").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado o no está aprobado"})
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
	Estado               bool               `json:"estado"`
	PrimeraImagenGaleria string             `json:"primera_imagen_galeria,omitempty"`
}

// GetMaterialsSummary lista resumen SOLO de materiales aprobados
func GetMaterialsSummary(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var materials []models.Material
	if err := db.Where("estado = ?", true).
		Preload("Galeria").
		Find(&materials).Error; err != nil {
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
			Estado:               m.Estado,
			PrimeraImagenGaleria: primeraImagen,
		})
	}

	c.JSON(http.StatusOK, summaries)
}

// GetMaterialsAdmin lista TODOS los materiales - Solo Admin
func GetMaterialsAdmin(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var materials []models.Material
	if err := db.Preload("Creador").
		Preload("Colaboradores").
		Preload("Pasos").
		Preload("Galeria").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando materiales: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, materials)
}

// GetMaterialAdmin obtiene un material por ID sin filtro de estado - Solo Admin
func GetMaterialAdmin(c *gin.Context) {
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
	if err := db.Where("id = ?", id).
		Preload("Creador").
		Preload("Colaboradores").
		Preload("Pasos").
		Preload("Galeria").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	c.JSON(http.StatusOK, material)
}

// GetMaterialsSummaryAdmin lista resumen de TODOS los materiales - Solo Admin
func GetMaterialsSummaryAdmin(c *gin.Context) {
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
			Estado:               m.Estado,
			PrimeraImagenGaleria: primeraImagen,
		})
	}

	c.JSON(http.StatusOK, summaries)
}

// GetMaterialsPendientes lista materiales pendientes de aprobación - Solo Admin
func GetMaterialsPendientes(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var materials []models.Material
	if err := db.Where("estado = ?", false).
		Preload("Creador").
		Preload("Galeria").
		Preload("Pasos").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando materiales pendientes: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":      len(materials),
		"materiales": materials,
	})
}

// GetDerivedMaterials obtiene los materiales que se derivan de un ID específico (PUNTO 3)
func GetDerivedMaterials(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	idStr := c.Param("id")
	parentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID padre inválido"})
		return
	}

	// Buscamos materiales donde derivado_de == parentID
	// Asumimos que solo mostramos los APROBADOS públicamente
	var derivedMaterials []models.Material
	if err := db.Where("derivado_de = ? AND estado = ?", parentID, true).
		Preload("Creador").
		Preload("Galeria").
		Find(&derivedMaterials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error buscando derivados: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, derivedMaterials)
}

// GetMaterialFilters obtiene listas únicas de herramientas y composiciones para filtros
func GetMaterialFilters(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var herramientas []string
	var composiciones []string

	// Obtener herramientas únicas y normalizadas
	err = db.Raw(`
		SELECT DISTINCT INITCAP(element)
		FROM materials, jsonb_array_elements_text(herramientas) AS element
		WHERE estado = true
		ORDER BY 1 ASC
	`).Scan(&herramientas).Error
	if err != nil {
		log.Printf("Error obteniendo filtros herramientas: %v", err)
	}

	// Obtener composiciones únicas y normalizadas
	err = db.Raw(`
		SELECT DISTINCT INITCAP(element)
		FROM materials, jsonb_array_elements_text(composicion) AS element
		WHERE estado = true
		ORDER BY 1 ASC
	`).Scan(&composiciones).Error
	if err != nil {
		log.Printf("Error obteniendo filtros composicion: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"herramientas": herramientas,
		"composicion":  composiciones,
	})
}
