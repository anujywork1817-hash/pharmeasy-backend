package confi

import (
	"fmt"
	"log"
	"os"

	"pharmeasy-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=root dbname=pharmeasy port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Medicine{},
		&models.Coupon{},
		&models.CouponUsage{},
		&models.Order{},
		&models.OrderItem{},
		&models.Review{},
		&models.TrackingUpdate{},
	)
	if err != nil {
		log.Fatal("❌ AutoMigrate failed:", err)
	}

	DB = db
	fmt.Println("✅ Database connected!")

	var count int64
	db.Model(&models.Medicine{}).Count(&count)
	fmt.Println("💊 Total medicines in DB:", count)
}
