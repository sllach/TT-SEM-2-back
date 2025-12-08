package auth

import (
	"net/http"
	"strings"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
)

// UpdateUsuario actualiza un usuario
func UpdateUsuario(c *gin.Context) {
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

	// Parsear request
	type UpdateRequest struct {
		Nombre string `json:"nombre"`
		Email  string `json:"email"`
		Rol    string `json:"rol"`
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request inválido: " + err.Error()})
		return
	}

	// Validar que al menos un campo sea proporcionado
	if req.Nombre == "" && req.Email == "" && req.Rol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar al menos un campo para actualizar (nombre, email o rol)"})
		return
	}

	// Actualizar campos proporcionados
	if req.Nombre != "" {
		usuario.Nombre = req.Nombre
	}

	if req.Email != "" {
		// Validar formato de email
		if !strings.Contains(req.Email, "@") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email inválido"})
			return
		}
		usuario.Email = req.Email
	}

	if req.Rol != "" {
		// Validar y normalizar rol
		rolLower := strings.ToLower(req.Rol)
		if rolLower != "lector" && rolLower != "colaborador" && rolLower != "administrador" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":            "Rol inválido",
				"roles_permitidos": []string{"lector", "colaborador", "administrador"},
				"rol_recibido":     req.Rol,
			})
			return
		}
		// Guardar con la capitalización correcta
		switch rolLower {
		case "lector":
			usuario.Rol = "lector"
		case "colaborador":
			usuario.Rol = "colaborador"
		case "administrador":
			usuario.Rol = "administrador"
		}
	}

	// Guardar cambios
	if err := db.Save(&usuario).Error; err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email ya está en uso por otro usuario"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando usuario: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usuario actualizado exitosamente",
		"usuario": usuario,
	})
}
