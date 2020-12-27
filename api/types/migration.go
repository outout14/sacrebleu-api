package types

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//SQLMigrate : Launch the database migration (creation of tables)
func SQLMigrate(db *gorm.DB) {
	logrus.Info("SQL : Database migration launched")
	db.AutoMigrate(&Domain{})
	db.AutoMigrate(&Record{})
	db.AutoMigrate(&User{})
}
