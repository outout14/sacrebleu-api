package utils

import (
	"fmt"

	"github.com/outout14/sacrebleu-dns/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//SQLDatabase Initialize the (My)SQL Database
//Requires a conf struct
func SQLDatabase(conf *utils.Conf) *gorm.DB {
	logrus.WithFields(logrus.Fields{"database": conf.Database.Db, "driver": conf.Database.Type}).Infof("SQL : Connection to DB")
	//Connect to the Database

	var gormLogLevel logger.LogLevel

	//Set GORM log level based on conf AppMode
	if conf.AppMode != "production" {
		gormLogLevel = logger.Info
	} else {
		gormLogLevel = logger.Silent
	}

	if conf.Database.Type == "postgresql" {
		dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s database=%s sslmode=disable", conf.Database.Username, conf.Database.Password, conf.Database.IP, conf.Database.Port, conf.Database.Db)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(gormLogLevel),
		})
		utils.CheckErr(err)
		return db

	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.Database.Username, conf.Database.Password, conf.Database.IP, conf.Database.Port, conf.Database.Db)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	utils.CheckErr(err)
	return db

}
