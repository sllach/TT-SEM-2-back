package main

import (
	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/handlers/material"
	auth "TT-SEM-2-BACK/api/handlers/usuarios"
	"TT-SEM-2-BACK/api/middleware"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	var err error

	// Intentar conectar hasta 5 veces
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {

		_, err = database.OpenGormDB()
		if err == nil {
			break
		}

		log.Printf("❌ Intento %d/%d falló: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	if err != nil {
		log.Fatalf("❌ Error crítico: No se pudo conectar a la Base de Datos después de %d intentos: %v", maxRetries, err)
	}

	log.Println("✅ Base de datos conectada correctamente")

	// Configuraracion CORS
	corsConfig := cors.Config{
		AllowOrigins:     []string{"https://tt-sem-2-front.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// Modo Release
	if os.Getenv("PORT") != "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(cors.New(corsConfig))

	// ========== RUTAS PÚBLICAS ==========
	router.POST("/auth/register", auth.RegisterUserFromGoogle)

	// Leer Materiales
	router.GET("/materials", material.GetMaterials)
	router.GET("/materials/:id", material.GetMaterial)
	router.GET("/materials/:id/derived", material.GetDerivedMaterials)
	router.GET("/materials/filters", material.GetMaterialFilters)
	router.GET("/materials-summary", material.GetMaterialsSummary)
	router.GET("/users/:google_id/public", auth.GetPublicUserProfile)

	// ========== RUTAS PROTEGIDAS ==========
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// Rutas generales
		protected.GET("/me", auth.GetMe)
		protected.POST("/users/request-role", auth.RequestCollaboratorRole)

		// ========== RUTAS ADMINISTRADOR Y COLABORADOR ==========
		adminCollab := protected.Group("/")
		adminCollab.Use(middleware.RequireRole("administrador", "colaborador"))
		{
			adminCollab.POST("/materials", material.CreateMaterial)
			adminCollab.PUT("/materials/:id", material.UpdateMaterial)

			// Notificaciones
			adminCollab.GET("/notifications", auth.GetNotifications)
			adminCollab.PATCH("/notifications/:id/read", auth.MarkNotificationRead)
		}

		// ========== RUTAS SOLO ADMINISTRADOR ==========
		adminOnly := protected.Group("/")
		adminOnly.Use(middleware.RequireRole("administrador"))
		{
			// Usuarios
			adminOnly.GET("/users", auth.GetUsuarios)
			adminOnly.GET("/users/:google_id", auth.GetUsuario)
			adminOnly.PUT("/users/:google_id", auth.UpdateUsuario)
			adminOnly.DELETE("/users/:google_id", auth.DeleteUsuario)
			adminOnly.DELETE("/users/:google_id/hard", auth.HardDeleteUsuario)
			adminOnly.GET("/users/stats", auth.GetDashboardStats)

			// Materiales Pendientes y Moderación
			adminOnly.GET("/materials/pending", material.GetMaterialsPendientes)
			adminOnly.POST("/materials/:id/approve", material.ApproveMaterial)
			adminOnly.POST("/materials/:id/reject", material.RejectMaterial)
			adminOnly.DELETE("/materials/:id", material.DeleteMaterial)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Servidor v1.0 iniciado en el puerto %s", port)
	router.Run(":" + port)
}
