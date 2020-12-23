package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/outout14/sacrebleu-api/utils"
	"github.com/sirupsen/logrus"
)

//Initialize : Initialize the mux router and the routes
func (a *App) Initialize(conf *utils.Conf) {

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

//initializeRoutes : Add all HTTP routes of the API to the HHTP server
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/demo", a.getDemo).Methods("GET")
}

//Run : Start the HTTP Server
func (a *App) Run(conf *utils.Conf) {
	logrus.WithFields(logrus.Fields{"ip": conf.App.IP, "port": conf.App.Port}).Infof("SERVER : Started")
	addr := fmt.Sprintf("%s:%v", conf.App.IP, conf.App.Port)
	logrus.Fatal(http.ListenAndServe(addr, a.Router))
}
