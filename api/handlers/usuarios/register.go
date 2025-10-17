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

// Registra o actualiza el usuario basado en Google profile
func RegisterUserFromGoogle(c *gin.Context) {
	// Parsear request
	type RegisterRequest struct {
		AccessToken string `json:"access_token" binding:"required"`
		GoogleID    string `json:"google_id" binding:"required"`
		Nombre      string `json:"nombre" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
	}
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parseando/validando JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request inválido: " + err.Error()})
		return
	}

	// Obtener JWT secret
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("SUPABASE_JWT_SECRET no configurado")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno: clave JWT no configurada"})
		return
	}

	// Parsear y validar el JWT
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(req.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		log.Printf("Token inválido: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido: " + err.Error()})
		return
	}
	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token no válido (posiblemente expirado)"})
		return
	}

	// Log para debug
	log.Printf("Email fetch de JWT: %s", claims.Email)
	log.Printf("Email enviado en request: %s", req.Email)

	// Verificar que el email coincida
	if !strings.EqualFold(claims.Email, req.Email) {
		log.Printf("Mismatch: fetch %s vs req %s", claims.Email, req.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Info de usuario no coincide"})
		return
	}

	// Conectar a la DB
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Buscar si el usuario ya existe por GoogleID
	var usuario models.Usuario
	if err := db.Where("google_id = ?", req.GoogleID).First(&usuario).Error; err == nil {
		// Existe: Actualizar nombre/email si cambiaron
		updated := false
		if usuario.Nombre != req.Nombre {
			usuario.Nombre = req.Nombre
			updated = true
		}
		if usuario.Email != req.Email {
			usuario.Email = req.Email
			updated = true
		}
		if updated {
			if err := db.Save(&usuario).Error; err != nil {
				if strings.Contains(err.Error(), "unique constraint") {
					c.JSON(http.StatusConflict, gin.H{"error": "Email ya registrado con otro usuario"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando usuario: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Usuario actualizado exitosamente", "usuario": usuario})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Usuario ya registrado, sin cambios necesarios", "usuario": usuario})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error buscando usuario: " + err.Error()})
		return
	}

	// Crear nuevo
	usuario = models.Usuario{
		GoogleID: req.GoogleID,
		Nombre:   req.Nombre,
		Email:    req.Email,
		Rol:      "user",
	}
	if err := db.Create(&usuario).Error; err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			if strings.Contains(err.Error(), "google_id") {
				c.JSON(http.StatusConflict, gin.H{"error": "Google ID ya registrado"})
			} else if strings.Contains(err.Error(), "email") {
				c.JSON(http.StatusConflict, gin.H{"error": "Email ya registrado"})
			} else {
				c.JSON(http.StatusConflict, gin.H{"error": "Registro duplicado"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando usuario: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado exitosamente", "usuario": usuario})
}
