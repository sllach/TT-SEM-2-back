package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Claims estructura para parsear el JWT de Supabase
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RegisterUserFromGoogle registra o actualiza el usuario y asegura el SupabaseID
func RegisterUserFromGoogle(c *gin.Context) {
	// 1. Parsear request
	type RegisterRequest struct {
		AccessToken string `json:"access_token" binding:"required"`
		GoogleID    string `json:"google_id" binding:"required"`
		Nombre      string `json:"nombre" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
	}
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parseando JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request inválido: " + err.Error()})
		return
	}

	// 2. Validar Token JWT
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("❌ SUPABASE_JWT_SECRET no configurado")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de configuración del servidor"})
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(req.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		log.Printf("❌ Token inválido: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// claims.Subject CONTIENE EL UUID DE SUPABASE (Ej: "737c9ec...")
	supabaseUID := claims.Subject

	// 3. Verificar consistencia de emails
	if !strings.EqualFold(claims.Email, req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El email del token no coincide con el usuario"})
		return
	}

	// 4. Conectar a DB
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de conexión BD"})
		return
	}

	// 5. Lógica de Upsert (Actualizar o Crear)
	var usuario models.Usuario

	// Buscamos si ya existe por GoogleID
	if err := db.Where("google_id = ?", req.GoogleID).First(&usuario).Error; err == nil {
		updated := false

		// Actualizamos campos básicos
		if usuario.Nombre != req.Nombre {
			usuario.Nombre = req.Nombre
			updated = true
		}
		if usuario.Email != req.Email {
			usuario.Email = req.Email
			updated = true
		}

		// Si el usuario existía pero no tenía SupabaseID (o cambió), lo actualizamos.
		if usuario.SupabaseID != supabaseUID {
			log.Printf("🔧 Reparando SupabaseID para usuario %s: %s -> %s", usuario.Email, usuario.SupabaseID, supabaseUID)
			usuario.SupabaseID = supabaseUID
			updated = true
		}

		if updated {
			if err := db.Save(&usuario).Error; err != nil {
				log.Printf("Error actualizando usuario: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando datos"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Usuario actualizado", "usuario": usuario})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Usuario sin cambios", "usuario": usuario})
		return

	} else if err != gorm.ErrRecordNotFound {
		// Error real de BD
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error buscando usuario"})
		return
	}

	// === USUARIO NUEVO: CREAR ===
	usuario = models.Usuario{
		SupabaseID: supabaseUID,
		GoogleID:   req.GoogleID,
		Nombre:     req.Nombre,
		Email:      req.Email,
		Rol:        "lector", // Rol por defecto
	}

	if err := db.Create(&usuario).Error; err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "Usuario ya existe (duplicado)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando usuario"})
		return
	}

	log.Printf("✅ Nuevo usuario registrado: %s (SupabaseID: %s)", usuario.Email, usuario.SupabaseID)
	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado exitosamente", "usuario": usuario})
}
