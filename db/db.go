package db

import (
	"fuzzyTest/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	
)

var gdb *gorm.DB

func Connect() {
	db, err := gorm.Open(sqlite.Open("zpath/db/test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect with DB", err)
	}
	gdb = db
	log.Print("Connected with DB successfully")
}

func GetGdb() *gorm.DB {
	return gdb
}

func Migrate() {
	err := gdb.AutoMigrate(
		&models.Item{},
		&models.Vendor{},
		&models.VendorItem{},
		&models.VendorOffer{},
	)
	
	if err != nil {
		log.Fatal("migrate failed", err)
	}
	log.Println("migration succeeded")
}
