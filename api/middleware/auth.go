package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims estructura para parsear el JWT de Supabase
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware valida el JWT de Supabase
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
		if tokenStr == authHeader { // No es Bearer
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		// Parsea y valida el token
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

		// Agrega el user ID o email al contexto para usarlo en handlers
		ctx := context.WithValue(c.Request.Context(), "user_id", claims.Subject) // claims.Subject es el UUID del user en Supabase
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		c.Request = c.Request.WithContext(ctx)

		c.Next() // Continúa al handler
	}
}
