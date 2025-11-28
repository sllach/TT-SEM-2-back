package auth

import (
	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStats struct {
	Pendientes int64 `json:"pendientes"`
	Aprobados  int64 `json:"aprobados"`
	Usuarios   int64 `json:"usuarios"`
}

func GetDashboardStats(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de conexión"})
		return
	}

	var stats DashboardStats

	// Hacemos los conteos directamente en la BD (Es mil veces más rápido)
	// 1. Contar pendientes
	db.Model(&models.Material{}).Where("estado = ?", false).Count(&stats.Pendientes)

	// 2. Contar aprobados
	db.Model(&models.Material{}).Where("estado = ?", true).Count(&stats.Aprobados)

	// 3. Contar usuarios
	db.Model(&models.Usuario{}).Count(&stats.Usuarios)

	c.JSON(http.StatusOK, stats)
}
