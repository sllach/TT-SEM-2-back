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

	// 1. Leer el cuerpo (Body) para ver si hay razón de rechazo
	var req DeleteRequest
	// Usamos ShouldBindJSON para que no falle si no envían nada (opcional)
	if err := c.ShouldBindJSON(&req); err != nil {
		// Si el JSON está mal formado o no existe, simplemente seguimos sin razón
		log.Println("No se envió razón de eliminación o JSON inválido")
	}

	// Verificar si el material existe
	var material models.Material
	if err := db.First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	// Guardamos datos temporales para la notificación
	creadorID := material.CreadorID
	nombreMaterial := material.Nombre

	// Eliminar asociaciones de colaboradores usando GORM
	if err := db.Model(&material).Association("Colaboradores").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando colaboradores: " + err.Error()})
		return
	}

	// Eliminar galería
	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.GaleriaMaterial{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando galería: " + err.Error()})
		return
	}

	// Eliminar pasos
	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PasoMaterial{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando pasos: " + err.Error()})
		return
	}

	// Eliminar propiedades mecánicas
	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PropiedadesMecanicas{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando propiedades mecánicas: " + err.Error()})
		return
	}

	// Eliminar propiedades perceptivas
	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PropiedadesPerceptivas{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando propiedades perceptivas: " + err.Error()})
		return
	}

	// Eliminar propiedades emocionales
	if err := db.Where("material_id = ?", id).Unscoped().Delete(&models.PropiedadesEmocionales{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando propiedades emocionales: " + err.Error()})
		return
	}

	// Eliminar el material principal con Unscoped para hard delete
	if err := db.Unscoped().Delete(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando material: " + err.Error()})
		return
	}

	// Solo enviamos notificación si el que borra NO es el dueño del material
	sendDeleteNotification(creadorID, nombreMaterial, req.Razon)

	c.JSON(http.StatusOK, gin.H{"message": "Material eliminado exitosamente"})
}

// Estructura para recibir la razón desde el frontend
type DeleteRequest struct {
	Razon string `json:"razon"`
}

// Helper específico para notificación de eliminación
func sendDeleteNotification(usuarioId string, materialName string, mensajeExtra string) {
	go func(uID string, mNombre string, extra string) {
		asyncDB, err := database.OpenGormDB()
		if err != nil {
			log.Printf("⚠️ Error conectando DB para notificación: %v", err)
			return
		}

		notifID := uuid.New()

		// Configuración del mensaje
		titulo := "Material Eliminado"
		mensaje := "El material '" + mNombre + "' ha sido eliminado del sistema."

		if extra != "" {
			mensaje += " Motivo: " + extra
		}

		// Tipo "info" (azul) o "rechazo" (rojo) según prefieras
		tipo := "info"

		nuevaNotif := models.Notificacion{
			ID:         notifID,
			UsuarioID:  uID,
			MaterialID: nil, // NIL: Porque el material ya no existe en la BD
			Titulo:     titulo,
			Mensaje:    mensaje,
			Tipo:       tipo,
			Link:       "/notification/#" + notifID.String(),
			Leido:      false,
		}

		if err := asyncDB.Create(&nuevaNotif).Error; err != nil {
			log.Printf("⚠️ Error guardando notificación de borrado: %v", err)
		} else {
			log.Printf("🗑️ Notificación de borrado enviada a %s", uID)
		}
	}(usuarioId, materialName, mensajeExtra)
}
