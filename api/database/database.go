package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"TT-SEM-2-BACK/api/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Variables globales para el Singleton
var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
	dbErr      error
)

// GetDB devuelve la instancia ÚNICA de la base de datos
func GetDB() (*gorm.DB, error) {
	dbOnce.Do(func() {
		dsn := config.DBURL()

		if dsn == "" {
			dbErr = fmt.Errorf("no se pudo generar la connection string")
			return
		}

		// Configuración de GORM
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
			PrepareStmt: true,
		})

		if err != nil {
			dbErr = fmt.Errorf("error conectando a la base de datos: %w", err)
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			dbErr = fmt.Errorf("error obteniendo sqlDB: %w", err)
			return
		}

		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(10 * time.Minute)

		if err := sqlDB.Ping(); err != nil {
			dbErr = fmt.Errorf("error haciendo ping a la DB: %w", err)
			return
		}

		log.Println("✅ Conexión a base de datos establecida")
		dbInstance = db
	})

	return dbInstance, dbErr
}

// OpenGormDB se mantiene por compatibilidad, pero ahora usa GetDB internamente.
func OpenGormDB() (*gorm.DB, error) {
	return GetDB()
}
