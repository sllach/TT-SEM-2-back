package auth

import (
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/middleware"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
)

// GetUsuarios lista todos los usuarios (solo admin)
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

// GetUsuario obtiene un usuario por GoogleID (solo admin)
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

// GetMe obtiene la información del usuario autenticado CON sus materiales
// Disponible para cualquier usuario autenticado
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

	// Obtener información del usuario
	var usuario models.Usuario
	if err := db.Where("google_id = ?", googleID).First(&usuario).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Obtener materiales creados por el usuario
	var materialesCreados []models.Material
	if err := db.Where("creador_id = ?", googleID).
		Preload("Galeria").
		Preload("Colaboradores").
		Preload("Pasos").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		Find(&materialesCreados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo materiales creados: " + err.Error()})
		return
	}

	// Obtener materiales donde es colaborador
	var materialesColaboracion []models.Material
	if err := db.Joins("JOIN colaboradores_material ON colaboradores_material.material_id = materials.id").
		Where("colaboradores_material.usuario_id = ?", googleID).
		Preload("Creador").
		Preload("Galeria").
		Preload("Colaboradores").
		Preload("PropiedadesMecanicas").
		Preload("PropiedadesPerceptivas").
		Preload("PropiedadesEmocionales").
		Find(&materialesColaboracion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo colaboraciones: " + err.Error()})
		return
	}

	// Contar materiales por estado
	var materialesAprobados, materialesPendientes int64
	db.Model(&models.Material{}).Where("creador_id = ? AND estado = ?", googleID, true).Count(&materialesAprobados)
	db.Model(&models.Material{}).Where("creador_id = ? AND estado = ?", googleID, false).Count(&materialesPendientes)

	c.JSON(http.StatusOK, gin.H{
		"usuario": gin.H{
			"nombre": usuario.Nombre,
			"email":  usuario.Email,
			"rol":    usuario.Rol,
		},
		"estadisticas": gin.H{
			"materiales_creados":    len(materialesCreados),
			"materiales_aprobados":  materialesAprobados,
			"materiales_pendientes": materialesPendientes,
			"colaboraciones":        len(materialesColaboracion),
		},
		"materiales_creados":     materialesCreados,
		"materiales_colaborador": materialesColaboracion,
	})
}

// GetUsuarioMateriales obtiene todos los materiales creados por un usuario (solo admin)
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
		Preload("Colaboradores").
		Preload("Pasos").
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

func GetPublicUserProfile(c *gin.Context) {
	googleID := c.Param("google_id")

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

	// Obtener SOLO materiales APROBADOS creados por el usuario (públicos)
	var materialesCreados []models.Material
	if err := db.Where("creador_id = ? AND estado = ?", googleID, true).
		Preload("Galeria").
		Preload("Colaboradores").
		Find(&materialesCreados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo materiales creados: " + err.Error()})
		return
	}

	// Obtener materiales APROBADOS donde es colaborador
	var materialesColaboracion []models.Material
	if err := db.Joins("JOIN colaboradores_material ON colaboradores_material.material_id = materials.id").
		Where("colaboradores_material.usuario_id = ? AND materials.estado = ?", googleID, true).
		Preload("Creador").
		Preload("Galeria").
		Preload("Colaboradores").
		Find(&materialesColaboracion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo colaboraciones: " + err.Error()})
		return
	}

	// Construir respuesta con información pública solamente
	c.JSON(http.StatusOK, gin.H{
		"perfil": gin.H{
			"google_id": usuario.GoogleID,
			"nombre":    usuario.Nombre,
			"rol":       usuario.Rol,
		},
		"estadisticas": gin.H{
			"materiales_creados": len(materialesCreados),
			"colaboraciones":     len(materialesColaboracion),
			"total_materiales":   len(materialesCreados) + len(materialesColaboracion),
		},
		"materiales_creados":     materialesCreados,
		"materiales_colaborador": materialesColaboracion,
	})
}
