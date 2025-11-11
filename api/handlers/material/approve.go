package material

import (
	"log"
	"net/http"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/middleware"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ApproveMaterial aprueba un material cambiando estado a true
func ApproveMaterial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Buscar el material
	var material models.Material
	if err := db.Preload("Creador").First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	// Verificar si ya está aprobado
	if material.Estado {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "El material ya está aprobado",
			"material": material,
		})
		return
	}

	// Aprobar el material
	material.Estado = true
	if err := db.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error aprobando material: " + err.Error()})
		return
	}

	// Log de la aprobación
	adminGoogleID, _ := middleware.GetUserGoogleID(c)
	log.Printf("✅ Material aprobado: %s (%s) por admin: %s", material.Nombre, material.ID, adminGoogleID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Material aprobado exitosamente",
		"material": material,
	})
}

// RejectMaterial rechaza/desaprueba un material cambiando estado a false
func RejectMaterial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Buscar el material
	var material models.Material
	if err := db.Preload("Creador").First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	// Verificar si ya está rechazado
	if !material.Estado {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "El material ya está rechazado/pendiente",
			"material": material,
		})
		return
	}

	// Rechazar/desaprobar el material
	material.Estado = false
	if err := db.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error rechazando material: " + err.Error()})
		return
	}

	// Log del rechazo
	adminGoogleID, _ := middleware.GetUserGoogleID(c)
	log.Printf("❌ Material rechazado: %s (%s) por admin: %s", material.Nombre, material.ID, adminGoogleID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Material rechazado/desaprobado exitosamente",
		"material": material,
	})
}

// ToggleApprovalMaterial cambia el estado de aprobación
func ToggleApprovalMaterial(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Buscar el material
	var material models.Material
	if err := db.Preload("Creador").First(&material, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Material no encontrado"})
		return
	}

	// Cambiar estado
	nuevoEstado := !material.Estado
	material.Estado = nuevoEstado

	if err := db.Save(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cambiando estado: " + err.Error()})
		return
	}

	// Log del cambio
	adminGoogleID, _ := middleware.GetUserGoogleID(c)
	estadoTexto := "rechazado"
	emoji := "❌"
	if nuevoEstado {
		estadoTexto = "aprobado"
		emoji = "✅"
	}
	log.Printf("%s Material %s: %s (%s) por admin: %s", emoji, estadoTexto, material.Nombre, material.ID, adminGoogleID)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Estado de aprobación cambiado exitosamente",
		"nuevo_estado": nuevoEstado,
		"material":     material,
	})
}
