// @title Sacrebleu DNS Server API
// @version 0.1
// @description This API allows you to manage a Sacrebleu DNS Database

// @contact.name Project github
// @contact.url https://github.com/outout14/sacrebleu-api

// @license.name AGPL-3.0
// @license.url https://raw.githubusercontent.com/outout14/sacrebleu-api/master/LICENCE

// @host localhost:5001
// @BasePath /api/
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-access-token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

package main

import (
	"flag"
	"fmt"

	"github.com/outout14/sacrebleu-api/api"
	"github.com/outout14/sacrebleu-api/api/types"
	"github.com/outout14/sacrebleu-dns/utils"
	"github.com/sirupsen/logrus"

	"github.com/outout14/sacrebleu-api/docs" //Swagger

	"gopkg.in/ini.v1"
)

var conf *utils.Conf

func main() {
	//Get the config patch from --config flag
	configPatch := flag.String("config", "extra/config.ini.example", "the patch to the config file") //Get the config patch from --config flag
	sqlMigration := flag.Bool("sqlmigrate", false, "initialize / migrate the database")              //Detect if migration asked
	adminCreate := flag.Bool("createadmin", false, "create admin user in the database")
	flag.Parse()

	//Load the INI configuration file
	conf = new(utils.Conf)
	err := ini.MapTo(conf, *configPatch)
	utils.CheckErr(err)

	//Swagger
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%v", conf.App.IP, conf.App.Port)

	//Set up the Logrus logger
	utils.InitLogger(conf)

	db := utils.SQLDatabase(conf)
	if *sqlMigration {
		types.SQLMigrate(db)
	}

	if *adminCreate {
		logrus.Warning("NEW ADMIN CREATION :")
		var newAdmin types.User
		fmt.Print("Enter the email :")
		fmt.Scanf("%s", &newAdmin.Email)
		fmt.Print("Enter the username :")
		fmt.Scanf("%s", &newAdmin.Username)
		fmt.Print("Enter the password :")
		fmt.Scanf("%s", &newAdmin.Password)
		newAdmin.IsAdmin = true
		newAdmin.Token = api.GenerateToken(newAdmin.Email)
		newAdmin.Password, _ = api.HashPassword(newAdmin.Password)
		err := newAdmin.CreateUser(db)
		if err == nil {
			logrus.Warningf("User created.")
			logrus.Warningf("NEW USER TOKEN : %s\n", newAdmin.Token)
		} else {
			logrus.Errorf("Can't create user : %s", err)
		}
		return
	}

	a := api.Server{DB: db}
	a.Initialize(conf)

	a.Run(conf)
}
