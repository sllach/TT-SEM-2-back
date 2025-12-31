package health

import (
	"net/http"
	"time"

	"TT-SEM-2-BACK/api/database"

	"github.com/gin-gonic/gin"
)

// CheckResponse define la estructura de nuestra respuesta de salud
type CheckResponse struct {
	Status    string    `json:"status"`    // "ok" o "error"
	Database  string    `json:"database"`  // "connected" o "disconnected"
	Timestamp time.Time `json:"timestamp"` // Hora del servidor
	Service   string    `json:"service"`   // Nombre del servicio
}

// HealthCheck verifica la conectividad del servidor y la base de datos
func HealthCheck(c *gin.Context) {
	dbStatus := "connected"
	httpStatus := http.StatusOK

	// 1. Obtener la instancia de GORM
	gormDB, err := database.GetDB()

	// Si falló la obtención de la instancia (ej: connection string vacía)
	if err != nil {
		dbStatus = "disconnected"
		httpStatus = http.StatusServiceUnavailable
	} else {
		// 2. Obtener el objeto genérico de base de datos SQL para hacer Ping
		sqlDB, err := gormDB.DB()
		if err != nil {
			dbStatus = "disconnected"
			httpStatus = http.StatusServiceUnavailable
		} else {
			// 3. Hacer Ping real a la base de datos
			if err := sqlDB.Ping(); err != nil {
				dbStatus = "disconnected"
				httpStatus = http.StatusServiceUnavailable
			}
		}
	}

	// 4. Construir respuesta
	response := CheckResponse{
		Status:    "ok",
		Database:  dbStatus,
		Timestamp: time.Now(),
		Service:   "bio-materials-api", // Nombre adaptado al proyecto actual
	}

	// Si la BD falla, cambiamos el estado general a error
	if dbStatus == "disconnected" {
		response.Status = "error"
	}

	c.JSON(httpStatus, response)
}
