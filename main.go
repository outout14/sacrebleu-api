package main

import (
	"flag"

	"github.com/outout14/sacrebleu-api/api"
	"github.com/outout14/sacrebleu-api/utils"

	"gopkg.in/ini.v1"
)

var conf *utils.Conf

func main() {
	//Get the config patch from --config flag
	configPatch := flag.String("config", "extra/config.ini.example", "the patch to the config file") //Get the config patch from --config flag
	sqlMigration := flag.Bool("sqlmigrate", false, "initialize / migrate the database")              //Detect if migration asked
	flag.Parse()

	//Load the INI configuration file
	conf = new(utils.Conf)
	err := ini.MapTo(conf, *configPatch)
	utils.CheckErr(err)

	//Set up the Logrus logger
	utils.InitLogger(conf)

	db := utils.SQLDatabase(conf)
	if *sqlMigration {
		utils.SQLMigrate(db)
	}

	a := api.App{DB: db}
	a.Initialize(conf)

	a.Run(conf)
}
