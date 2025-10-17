package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"TT-SEM-2-BACK/api/database"
	"TT-SEM-2-BACK/api/models"

	"github.com/gin-gonic/gin"
)

// Estructura del user de Supabase
type SupabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// RegisterUserFromGoogle registra o actualiza el usuario basado en Google profile
func RegisterUserFromGoogle(c *gin.Context) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		log.Println("Variables de Supabase no configuradas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}

	// Parsear request
	type RegisterRequest struct {
		AccessToken string `json:"access_token"`
		GoogleID    string `json:"google_id"`
		Nombre      string `json:"nombre"`
		Email       string `json:"email"`
	}
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request inválido: " + err.Error()})
		return
	}

	// Validar el access_token con Supabase API directamente
	apiURL := fmt.Sprintf("%s/auth/v1/user", supabaseURL)
	httpReq, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Error creando request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	httpReq.Header.Set("apikey", supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Error enviando request a Supabase: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validando token"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Error en respuesta de Supabase: %d - %s", resp.StatusCode, string(body))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// Parsear la respuesta
	var supabaseResp struct {
		User SupabaseUser `json:"user"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error leyendo respuesta"})
		return
	}
	if err := json.Unmarshal(body, &supabaseResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parseando respuesta: " + err.Error()})
		return
	}

	// Verificar que el user de Supabase coincida con la info enviada
	if supabaseResp.User.Email != req.Email {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Info de usuario no coincide"})
		return
	}

	// Conectar a tu DB
	db, err := database.OpenGormDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error conectando a la DB"})
		return
	}

	// Buscar si el usuario ya existe por GoogleID
	var usuario models.Usuario
	if err := db.Where("google_id = ?", req.GoogleID).First(&usuario).Error; err == nil {
		// Existe: Actualizar nombre/email si cambiaron (opcional)
		usuario.Nombre = req.Nombre
		usuario.Email = req.Email
		if err := db.Save(&usuario).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando usuario"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Usuario actualizado", "usuario": usuario})
		return
	}

	// No existe: Crear nuevo
	usuario = models.Usuario{
		GoogleID: req.GoogleID,
		Nombre:   req.Nombre,
		Email:    req.Email,
		Rol:      "user", // Default
	}
	if err := db.Create(&usuario).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando usuario: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario registrado", "usuario": usuario})
}
