package auth

import (
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/middleware"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
)

// GetUsuarios lista todos los usuarios
func GetUsuarios(c *gin.Context) {
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var usuarios []models.Usuario
	if err := db.Find(&usuarios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando usuarios: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, usuarios)
}

// GetUsuario obtiene un usuario por GoogleID
func GetUsuario(c *gin.Context) {
	googleID := c.Param("google_id")

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var usuario models.Usuario
	if err := db.Where("google_id = ?", googleID).First(&usuario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, usuario)
}

// GetMe obtiene la información del usuario autenticado
func GetMe(c *gin.Context) {
	googleID, exists := middleware.GetUserGoogleID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Datos de usuario incompletos"})
		return
	}

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	var usuario models.Usuario
	if err := db.Where("google_id = ?", googleID).First(&usuario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, usuario)
}

// GetUsuarioMateriales obtiene todos los materiales creados por un usuario
func GetUsuarioMateriales(c *gin.Context) {
	googleID := c.Param("google_id")

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Verificar que el usuario existe
	var usuario models.Usuario
	if err := db.Where("google_id = ?", googleID).First(&usuario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Obtener materiales del usuario
	var materiales []models.Material
	if err := db.Where("creador_id = ?", googleID).
		Preload("Galeria").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		Find(&materiales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo materiales: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usuario": gin.H{
			"google_id": usuario.GoogleID,
			"nombre":    usuario.Nombre,
			"email":     usuario.Email,
			"rol":       usuario.Rol,
		},
		"materiales": materiales,
		"total":      len(materiales),
	})
}
