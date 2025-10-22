package main

import (
	"TT-SEM-2-BACK/api/config"
	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/handlers/material"
	auth "TT-SEM-2-BACK/api/handlers/usuarios"
	"TT-SEM-2-BACK/api/middleware"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.OpenGormDB()
	if err != nil {
		log.Fatalf("Error al conectarse a la Base de Datos: %v", err)
	}

	db.AutoMigrate(
	//&models.Usuario{},
	//&models.Material{},
	//&models.PropiedadesEmocionales{},
	//&models.PropiedadesMecanicas{},
	//&models.PropiedadesPerceptivas{},
	//&models.PasoMaterial{},
	//&models.GaleriaMaterial{},
	//&models.ColaboradorMaterial{},
	)

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

	// ========== RUTAS PÚBLICAS ==========
	router.POST("/auth/register", auth.RegisterUserFromGoogle)

	// Leer
	router.GET("/materials", material.GetMaterials)
	router.GET("/materials/:id", material.GetMaterial)
	router.GET("/materials-summary", material.GetMaterialsSummary)

	// ========== RUTAS PROTEGIDAS ==========
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// ========== RUTAS SOLO PARA ADMINISTRADOR Y COLABORADOR ==========
		adminCollab := protected.Group("/")
		adminCollab.Use(middleware.RequireRole("administrador", "colaborador"))
		{
			// Crear material
			adminCollab.POST("/materials", material.CreateMaterial)

			// Actualizar material
			adminCollab.PUT("/materials/:id", material.UpdateMaterial)
		}

		// ========== RUTAS SOLO PARA ADMINISTRADOR ==========
		adminOnly := protected.Group("/")
		adminOnly.Use(middleware.RequireRole("administrador"))
		{
			// Leer
			adminOnly.GET("/users", auth.GetUsuarios)           // Listar todos los usuarios
			adminOnly.GET("/users/:google_id", auth.GetUsuario) // Obtener un usuario específico

			// Actualizar
			adminOnly.PUT("/users/:google_id", auth.UpdateUsuario) //Actualizar Usuario

			// Eliminar
			adminOnly.DELETE("/materials/:id", material.DeleteMaterial)        // Eliminar material
			adminOnly.DELETE("/users/:google_id", auth.DeleteUsuario)          // Eliminar usuario (soft delete)
			adminOnly.DELETE("/users/:google_id/hard", auth.HardDeleteUsuario) //Eliminar usuario (hard delete)

		}
	}

	log.Println("Servidor iniciado en :8080")

	router.Run(":8080")
}

