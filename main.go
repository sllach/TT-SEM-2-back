package main

import (
	"TT-SEM-2-BACK/api/config"
	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/handlers/material"
	auth "TT-SEM-2-BACK/api/handlers/usuarios"
	"TT-SEM-2-BACK/api/middleware"

	//"TT-SEM-2-BACK/api/models"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	var db *gorm.DB
	var err error

	// Intentar conectar hasta 5 veces
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = database.OpenGormDB()
		if err == nil {
			break
		}

		log.Printf("❌ Intento %d/%d falló: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	if err != nil {
		log.Fatalf("Error al conectarse a la Base de Datos después de %d intentos: %v", maxRetries, err)
	}

	log.Println("✅ Base de datos conectada")

	db.AutoMigrate(
	/*&models.Usuario{},
	&models.Material{},
	&models.PropiedadesEmocionales{},
	&models.PropiedadesMecanicas{},
	&models.PropiedadesPerceptivas{},
	&models.PasoMaterial{},
	&models.GaleriaMaterial{},
	&models.Notificacion{},*/
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
	router.GET("/users/:google_id/public", auth.GetPublicUserProfile) // Perfil público de usuario

	// ========== RUTAS PROTEGIDAS ==========
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{

		protected.GET("/me", auth.GetMe) // Listar Perfil del usuario juntos con sus materiales

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
			adminOnly.GET("/users", auth.GetUsuarios)                            // Listar todos los usuarios
			adminOnly.GET("/users/:google_id", auth.GetUsuario)                  // Obtener un usuario específico
			adminOnly.GET("/materials/pending", material.GetMaterialsPendientes) // Listar materiales pendientes de aprobación
			adminOnly.GET("/users/stats", auth.GetDashboardStats)                //Listar la cantidad de usuarios y  materiales

			// Actualizar
			adminOnly.PUT("/users/:google_id", auth.UpdateUsuario)             //Actualizar Usuario
			adminOnly.POST("/materials/:id/approve", material.ApproveMaterial) // Aprobar material
			adminOnly.POST("/materials/:id/reject", material.RejectMaterial)   // Rechazar material

			// Eliminar
			adminOnly.DELETE("/materials/:id", material.DeleteMaterial)        // Eliminar material
			adminOnly.DELETE("/users/:google_id", auth.DeleteUsuario)          // Eliminar usuario (soft delete)
			adminOnly.DELETE("/users/:google_id/hard", auth.HardDeleteUsuario) // Eliminar usuario (hard delete)

		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Escuchando en el puerto %s", port)
	router.Run(":" + port)
}
