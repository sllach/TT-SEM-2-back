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
	// NOTA: Ya no hacemos Preload de propiedades porque son columnas JSONB y se cargan solas.
	if err := db.Where("estado = ?", true).
		Preload("Creador").
		Preload("Colaboradores").
		Preload("Pasos").
		Preload("Galeria").
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
		First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado o no está aprobado"})
		return
	}

	c.JSON(http.StatusOK, material)
}

// SummaryMaterial
type SummaryMaterial struct {
	ID                   uuid.UUID              `json:"id"`
	Nombre               string                 `json:"nombre"`
	Descripcion          string                 `json:"descripcion"`
	Composicion          models.JSONComponentes `json:"composicion"`
	Herramientas         models.StringArray     `json:"herramientas"`
	DerivadoDe           uuid.UUID              `json:"derivado_de"`
	Estado               bool                   `json:"estado"`
	PrimeraImagenGaleria string                 `json:"primera_imagen_galeria,omitempty"`
}

// GetMaterialsSummary lista resumen SOLO de materiales aprobados
func GetMaterialsSummary(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var materials []models.Material
	// Solo necesitamos cargar Galería para la foto de portada
	if err := db.Where("estado = ?", true).
		Preload("Galeria").
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando resumen: " + err.Error()})
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
			Herramientas:         m.Herramientas,
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
		First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	c.JSON(http.StatusOK, material)
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
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando pendientes: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":      len(materials),
		"materiales": materials,
	})
}

// GetDerivedMaterials obtiene los materiales derivados
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

// GetMaterialFilters obtiene filtros (Herramientas y Composición)
func GetMaterialFilters(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var herramientas []string
	var composiciones []string

	// 1. Obtener herramientas únicas (Array de strings JSON)
	err = db.Raw(`
        SELECT DISTINCT INITCAP(element)
        FROM materials, jsonb_array_elements_text(herramientas) AS element
        WHERE estado = true
        ORDER BY 1 ASC
    `).Scan(&herramientas).Error
	if err != nil {
		log.Printf("Error obteniendo filtros herramientas: %v", err)
	}

	// 2. Obtener composiciones únicas (Array de Objetos JSON)
	err = db.Raw(`
        SELECT DISTINCT INITCAP(element->>'elemento')
        FROM materials, jsonb_array_elements(composicion) AS element
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
