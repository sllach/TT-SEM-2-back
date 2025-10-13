package material

import (
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeleteMaterial maneja la eliminación de un material
func DeleteMaterial(c *gin.Context) {
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

	// Verificar si el material existe
	var material models.Material
	if err := db.First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	// Eliminar relaciones dependientes con Unscoped para hard delete
	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.ColaboradorMaterial{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando colaboradores: " + err.Error()})
		return
	}

	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.GaleriaMaterial{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando galería: " + err.Error()})
		return
	}

	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PasoMaterial{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando pasos: " + err.Error()})
		return
	}

	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PropiedadesMecanicas{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando propiedades mecánicas: " + err.Error()})
		return
	}

	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PropiedadesPerceptivas{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando propiedades perceptivas: " + err.Error()})
		return
	}

	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PropiedadesEmocionales{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando propiedades emocionales: " + err.Error()})
		return
	}

	// Eliminar el material principal con Unscoped para hard delete
	if err := db.Unscoped().Delete(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando material: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Material eliminado exitosamente"})
}
