package middleware

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
)

// Claims estructura para parsear el JWT de Supabase
type Claims struct {
	jwt.RegisteredClaims
}

// AuthMiddleware valida el JWT de Supabase y carga el usuario local
func AuthMiddleware() gin.HandlerFunc {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("SUPABASE_JWT_SECRET no configurado")
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header requerido"})
			c.Abort()
			return
		}

		// Extrae el token (Bearer <token>)
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		// Parsear y validar el JWT
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido: " + err.Error()})
			c.Abort()
			return
		}

		// Conectar a DB para buscar usuario local por SupabaseID
		db, err := database.OpenGormDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
			c.Abort()
			return
		}

		var usuario models.Usuario
		if err := db.Where("supabase_id = ?", claims.Subject).First(&usuario).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no registrado en el sistema"})
			c.Abort()
			return
		}

		c.Set("rol", usuario.Rol)
		c.Set("google_id", usuario.GoogleID)

		c.Next()
	}
}

// RequireRole middleware que verifica si el usuario tiene uno de los roles permitidos
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolAny, exists := c.Get("rol")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
			c.Abort()
			return
		}

		userRole := strings.ToLower(rolAny.(string))

		// Verificar si el rol del usuario está en la lista de roles permitidos
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if userRole == strings.ToLower(allowedRole) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "No tienes permisos para realizar esta acción",
				"required_roles": allowedRoles,
				"your_role":      userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IsAdmin verifica si el usuario es administrador
func IsAdmin(c *gin.Context) bool {
	rolAny, exists := c.Get("rol")
	if !exists {
		return false
	}
	return strings.ToLower(rolAny.(string)) == "administrador"
}

// GetUserGoogleID obtiene el GoogleID del usuario autenticado
func GetUserGoogleID(c *gin.Context) (string, bool) {
	googleIDAny, exists := c.Get("google_id")
	if !exists {
		return "", false
	}
	return googleIDAny.(string), true
}
