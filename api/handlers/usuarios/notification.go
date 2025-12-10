package auth

import (
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/middleware"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
)

// GetNotifications: Obtiene todas las notificaciones del usuario
func GetNotifications(c *gin.Context) {
	// 1. Obtener el ID del usuario logueado (GoogleID)
	googleID, exists := middleware.GetUserGoogleID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// 2. Conectar a la BD
	db, err := database.OpenGormDB() // O database.GetDB(), según tu configuración
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de conexión a BD"})
		return
	}

	var notificaciones []models.Notificacion

	// 3. Query:
	// - Filtrar por UsuarioID
	// - Ordenar por fecha de creación descendente (más nuevas primero)
	result := db.Where("usuario_id = ?", googleID).
		Order("created_at desc").
		Find(&notificaciones)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error buscando notificaciones"})
		return
	}

	c.JSON(http.StatusOK, notificaciones)
}

// MarkNotificationRead: Marca una notificación como leída
func MarkNotificationRead(c *gin.Context) {
	// 1. Obtener ID del usuario
	googleID, exists := middleware.GetUserGoogleID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// 2. Obtener ID de la notificación desde la URL
	notifID := c.Param("id")

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de conexión BD"})
		return
	}

	// 3. Update seguro:
	// Solo actualiza si el ID coincide Y si pertenece al usuario actual (seguridad)
	res := db.Model(&models.Notificacion{}).
		Where("id = ? AND usuario_id = ?", notifID, googleID).
		Update("leido", true)

	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando notificación"})
		return
	}

	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notificación no encontrada o no te pertenece"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notificación leída"})
}
