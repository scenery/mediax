package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scenery/mediax/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB(migrate bool) {
	var err error
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:        time.Second,
			LogLevel:             logger.Error,
			ParameterizedQueries: true,
			Colorful:             true,
		},
	)

	dbPath := "mediax.db"

	_, err = os.Stat(dbPath)
	databaseExists := err == nil

	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		TranslateError: true,
		Logger:         dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if !databaseExists || migrate {
		err = db.AutoMigrate(&models.Subject{})
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		fmt.Println("Database migration successful.")
		if migrate {
			os.Exit(0)
		}
	}
}

func GetDB() *gorm.DB {
	return db
}
