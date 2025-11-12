package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	// Cargar .env solo si no estamos en producción (Render)
	if os.Getenv("RENDER") == "" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("⚠️ Advertencia: No se encontró el archivo .env o hubo un error al cargarlo.")
		} else {
			log.Println("✅ Archivo .env cargado exitosamente")
		}
	} else {
		log.Println("🌐 Ejecutando en Render - usando variables de entorno del sistema")
	}
}

func DBURL() string {
	DBUser := strings.TrimSpace(os.Getenv("DB_USER"))
	DBPassword := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	DBHost := strings.TrimSpace(os.Getenv("DB_HOST"))
	DBPort := strings.TrimSpace(os.Getenv("DB_PORT"))
	DBName := strings.TrimSpace(os.Getenv("DB_NAME"))

	if DBUser == "" || DBPassword == "" || DBHost == "" || DBPort == "" || DBName == "" {
		log.Printf("❌ ERROR: Variables de entorno incompletas")
		log.Printf("   DB_USER: '%s'", DBUser)
		log.Printf("   DB_HOST: '%s'", DBHost)
		log.Printf("   DB_PORT: '%s'", DBPort)
		log.Printf("   DB_NAME: '%s'", DBName)
		return ""
	}

	sslMode := "require"
	if os.Getenv("RENDER") == "" && os.Getenv("LOCAL_DEV") == "true" {
		sslMode = "disable"
	}

	// Session Pooler connection string
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		DBUser, DBPassword, DBHost, DBPort, DBName, sslMode,
	)

	log.Printf("🔗 Session Pooler: host=%s port=%s ssl=%s", DBHost, DBPort, sslMode)
	return connectionString
}

// GetEnv obtiene una variable de entorno con un valor por defecto
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
