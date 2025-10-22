package auth

import (
	"log"
	"net/http"
	"strings"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/middleware"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
)

// DeleteUsuario elimina un usuario (solo admin)
func DeleteUsuario(c *gin.Context) {
	googleID := c.Param("google_id")

	// Verificar que el admin no se esté intentando eliminar a sí mismo
	currentUserGoogleID, _ := middleware.GetUserGoogleID(c)
	if currentUserGoogleID == googleID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "No puedes eliminar tu propia cuenta",
			"detail": "Por seguridad, no puedes eliminar tu propia cuenta de administrador",
		})
		return
	}

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Buscar usuario
	var usuario models.Usuario
	if err := db.Where("google_id = ?", googleID).First(&usuario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Verificar cuántos materiales tiene el usuario
	var countMateriales int64
	if err := db.Model(&models.Material{}).Where("creador_id = ?", googleID).Count(&countMateriales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verificando materiales: " + err.Error()})
		return
	}

	// Log de la operación
	log.Printf("🗑️ Soft delete: Usuario %s (%s) - GoogleID: %s - Materiales: %d",
		usuario.Nombre, usuario.Email, usuario.GoogleID, countMateriales)

	// Soft delete (GORM automáticamente usa DeletedAt)
	if err := db.Delete(&usuario).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando usuario: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Usuario eliminado exitosamente (soft delete)",
		"detail":           "El usuario ha sido marcado como eliminado pero sus datos permanecen en la base de datos",
		"google_id":        googleID,
		"nombre":           usuario.Nombre,
		"email":            usuario.Email,
		"materiales_count": countMateriales,
	})
}

// HardDeleteUsuario elimina permanentemente un usuario (solo admin)
func HardDeleteUsuario(c *gin.Context) {
	googleID := c.Param("google_id")

	// Verificar que el admin no se esté intentando eliminar a sí mismo
	currentUserGoogleID, _ := middleware.GetUserGoogleID(c)
	if currentUserGoogleID == googleID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "No puedes eliminar tu propia cuenta",
			"detail": "Por seguridad, no puedes eliminar tu propia cuenta de administrador",
		})
		return
	}

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Buscar usuario (incluso si está soft deleted)
	var usuario models.Usuario
	if err := db.Unscoped().Where("google_id = ?", googleID).First(&usuario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Verificar si el usuario tiene materiales asociados
	var countMateriales int64
	if err := db.Model(&models.Material{}).Where("creador_id = ?", googleID).Count(&countMateriales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verificando materiales: " + err.Error()})
		return
	}

	if countMateriales > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":            "No se puede eliminar el usuario porque tiene materiales asociados",
			"detail":           "Este usuario ha creado materiales. Primero elimina o reasigna sus materiales",
			"materiales_count": countMateriales,
			"suggestion":       "Puedes usar soft delete (DELETE /usuarios/:google_id) para mantener la integridad de los datos",
		})
		return
	}

	// Verificar si es colaborador en otros materiales
	var countColaboraciones int64
	if err := db.Model(&models.ColaboradorMaterial{}).Where("usuario_id = ?", googleID).Count(&countColaboraciones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verificando colaboraciones: " + err.Error()})
		return
	}

	// Eliminar colaboraciones si existen
	if countColaboraciones > 0 {
		log.Printf("⚠️ Eliminando %d colaboraciones del usuario %s", countColaboraciones, googleID)
		if err := db.Where("usuario_id = ?", googleID).Delete(&models.ColaboradorMaterial{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando colaboraciones: " + err.Error()})
			return
		}
	}

	// ADVERTENCIA
	log.Printf("⚠️⚠️⚠️ HARD DELETE: Eliminando permanentemente usuario %s (%s) - GoogleID: %s",
		usuario.Nombre, usuario.Email, usuario.GoogleID)

	// Hard delete
	if err := db.Unscoped().Delete(&usuario).Error; err != nil {
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates foreign key constraint") {
			c.JSON(http.StatusConflict, gin.H{
				"error":  "No se puede eliminar el usuario debido a restricciones de integridad referencial",
				"detail": "El usuario tiene registros relacionados en otras tablas que no pueden ser eliminados automáticamente",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando usuario: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                   "Usuario eliminado permanentemente de la base de datos",
		"google_id":                 googleID,
		"nombre":                    usuario.Nombre,
		"email":                     usuario.Email,
		"colaboraciones_eliminadas": countColaboraciones,
		"warning":                   "⚠️ Esta acción no se puede deshacer",
	})
}
