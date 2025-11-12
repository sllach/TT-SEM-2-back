package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenGormDB() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Usar sslmode=require para Supabase
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname,
	)

	// Configuración con timeouts
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	})

	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	// Configurar pool de conexiones
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configuraciones importantes para serverless
	sqlDB.SetMaxIdleConns(2)                   // Pocas conexiones idle
	sqlDB.SetMaxOpenConns(10)                  // Máximo de conexiones
	sqlDB.SetConnMaxLifetime(time.Hour)        // Reconectar cada hora
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Cerrar idle después de 10min

	// Ping para verificar conexión
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error haciendo ping a la DB: %w", err)
	}

	log.Println("✅ Conectado a la base de datos exitosamente")
	return db, nil
}
