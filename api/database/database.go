package database

import (
	"TT-SEM-2-BACK/api/config"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenGormDB() (*gorm.DB, error) {
	// Obtener connection string desde config
	dsn := config.DBURL()

	if dsn == "" {
		return nil, fmt.Errorf("no se pudo generar la connection string - verifica las variables de entorno")
	}

	// Configurar GORM - DESACTIVAR PrepareStmt para serverless
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: false, // ← CAMBIAR A FALSE para Render
	})

	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	// Configurar pool de conexiones
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo sqlDB: %w", err)
	}

	// Configuraciones optimizadas para serverless
	sqlDB.SetMaxIdleConns(1)                  // ← Reducir a 1
	sqlDB.SetMaxOpenConns(5)                  // ← Reducir a 5
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // ← Reducir a 5 minutos
	sqlDB.SetConnMaxIdleTime(2 * time.Minute) // ← Reducir a 2 minutos

	// Verificar conexión
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error haciendo ping a la DB: %w", err)
	}

	log.Println("✅ Conectado a la base de datos exitosamente")
	return db, nil
}
