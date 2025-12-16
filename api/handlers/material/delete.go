package material

import (
	"log"
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeleteMaterial maneja la eliminación de un material
func DeleteMaterial(c *gin.Context) {
	db, err := database.GetDB()
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

	// Leer razón de eliminación
	var req struct {
		Razon string `json:"razon"`
	}
	c.ShouldBindJSON(&req)

	// Verificar existencia
	var material models.Material
	if err := db.First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	creadorID := material.CreadorID
	nombreMaterial := material.Nombre

	// 1. Eliminar Colaboradores (Tabla Intermedia Explicita)
	// Usamos el nombre real de la tabla para evitar errores
	if err := db.Table("material_colaboradores").Where("material_id = ?", id).Delete(nil).Error; err != nil {
		log.Printf("⚠️ Error borrando colaboradores: %v", err)
		// No retornamos error fatal, intentamos seguir borrando lo demás
	}

	// 2. Eliminar Galería
	db.Where("material_id = ?", id).Unscoped().Delete(&models.GaleriaMaterial{})

	// 3. Eliminar Pasos
	db.Where("material_id = ?", id).Unscoped().Delete(&models.PasoMaterial{})

	// 4. Eliminar Material
	// (Las propiedades JSON se borran junto con el material, no hay que hacer nada extra)
	if err := db.Unscoped().Delete(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando material: " + err.Error()})
		return
	}

	// Notificar
	sendDeleteNotification(creadorID, nombreMaterial, req.Razon)

	c.JSON(http.StatusOK, gin.H{"message": "Material eliminado exitosamente"})
}

func sendDeleteNotification(usuarioId string, materialName string, mensajeExtra string) {
	go func() {
		db, _ := database.GetDB()
		msg := "El material '" + materialName + "' ha sido eliminado."
		if mensajeExtra != "" {
			msg += " Motivo: " + mensajeExtra
		}

		db.Create(&models.Notificacion{
			UsuarioID: usuarioId,
			Titulo:    "Material Eliminado",
			Mensaje:   msg,
			Tipo:      "info",
			Leido:     false,
		})
	}()
}
