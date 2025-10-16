package main

import (
	"TT-SEM-2-BACK/api/config"
	"TT-SEM-2-BACK/api/handlers/material"
	auth "TT-SEM-2-BACK/api/handlers/usuarios"

	//"TT-SEM-2-BACK/api/middleware"

	//	"TT-SEM-2-BACK/api/database"

	//	"TT-SEM-2-BACK/api/models"
	"fmt"
	//"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	/*db, err := database.OpenGormDB()
	if err != nil {
		log.Fatalf("Error al conectarse a la Base de Datos: %v", err)
	}

	db.AutoMigrate(
		&models.Usuario{},
		&models.Material{},
		&models.PropiedadesEmocionales{},
		&models.PropiedadesMecanicas{},
		&models.PropiedadesPerceptivas{},
		&models.PasoMaterial{},
		&models.GaleriaMaterial{},
		&models.ColaboradorMaterial{},
	)*/

	fmt.Print(config.DBURL())

	// Configurar CORS
	corsConfig := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router := gin.Default()
	router.Use(cors.New(corsConfig))

	router.POST("/auth/register", auth.RegisterUserFromGoogle)

	//Crear
	router.POST("/materials", material.CreateMaterial)

	//Leer
	router.GET("/materials", material.GetMaterials)
	router.GET("/materials/:id", material.GetMaterial)
	router.GET("/materials-summary", material.GetMaterialsSummary)

	//Actualizar
	router.PUT("/materials/:id", material.UpdateMaterial)

	//Eliminar
	router.DELETE("/materials/:id", material.DeleteMaterial)

	router.Run(":8080")
}
