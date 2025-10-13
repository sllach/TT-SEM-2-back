package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func DBURL() string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Advertencia: No se encontró el archivo .env o hubo un error al cargarlo.")
	}

	DBUser := strings.TrimSpace(os.Getenv("DB_USER"))
	DBPassword := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
	DBHost := strings.TrimSpace(os.Getenv("DB_HOST"))
	DBPort := strings.TrimSpace(os.Getenv("DB_PORT"))
	DBName := strings.TrimSpace(os.Getenv("DB_NAME"))

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&statement_cache_mode=describe", DBUser, DBPassword, DBHost, DBPort, DBName)
}
