package auth

import (
	"log"
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/middleware"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
)

// RequestCollaboratorRole: Usuario solicita ser colaborador
func RequestCollaboratorRole(c *gin.Context) {
	// 1. Obtener ID del usuario que solicita
	googleID, exists := middleware.GetUserGoogleID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de conexión a BD"})
		return
	}

	// 2. Obtener datos del usuario para el mensaje
	var solicitante models.Usuario
	if err := db.Where("google_id = ?", googleID).First(&solicitante).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Validación opcional: Si ya es colaborador o admin, no tiene sentido pedirlo
	if solicitante.Rol == "colaborador" || solicitante.Rol == "administrador" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ya tienes permisos de colaborador o superior"})
		return
	}

	// 3. Notificar a los Administradores (Asíncrono)
	go func() {
		// Buscar todos los admins
		var admins []models.Usuario
		if err := db.Where("rol = ?", "administrador").Find(&admins).Error; err != nil {
			log.Printf("⚠️ Error buscando admins para solicitud de rol: %v", err)
			return
		}

		// Crear notificación para cada uno
		for _, admin := range admins {
			notif := models.Notificacion{
				UsuarioID: admin.GoogleID,
				// No asociamos MaterialID porque es una solicitud de usuario
				MaterialID: nil,
				Titulo:     "Solicitud de Rol: Colaborador",
				Mensaje:    "El usuario " + solicitante.Nombre + " (" + solicitante.Email + ") solicita ser Colaborador.",
				Tipo:       "solicitud_rol", // Tipo especial para manejar íconos en el front
				Link:       "/admin",        // Link a la gestión de usuarios
				Leido:      false,
			}
			if err := db.Create(&notif).Error; err != nil {
				log.Printf("Error creando notificación para admin %s: %v", admin.Email, err)
			}
		}
		log.Printf("🔔 Solicitud de rol enviada a %d administradores.", len(admins))
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Solicitud enviada exitosamente. Un administrador revisará tu petición.",
	})
}
