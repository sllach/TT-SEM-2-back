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
	dsn := config.DBURL()

	if dsn == "" {
		return nil, fmt.Errorf("no se pudo generar la connection string")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Session pooler SÍ soporta prepared statements
	})

	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo sqlDB: %w", err)
	}

	// Configuración para session pooler
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error haciendo ping a la DB: %w", err)
	}

	log.Println("✅ Conectado a la base de datos (session pooler)")
	return db, nil
}
