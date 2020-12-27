package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/outout14/sacrebleu-api/docs" //Swagger
	"github.com/outout14/sacrebleu-dns/utils"
	"github.com/sirupsen/logrus"

	"github.com/swaggo/http-swagger"
)

//Initialize : Initialize the mux router and the routes
func (a *Server) Initialize(conf *utils.Conf) {
	a.Router = mux.NewRouter()
	a.APIRouter = a.Router.PathPrefix("/api").Subrouter()

	a.APIRouter.Use(JwtVerify(a))
	a.initializeRoutes()
}

//initializeRoutes : Add all HTTP routes of the API to the HHTP server
func (a *Server) initializeRoutes() {
	//Swagger doc
	logrus.Debug("[SERVER] Documentation Init")
	a.Router.PathPrefix("/").Handler(httpSwagger.Handler(
		httpSwagger.URL("./swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
	))
	logrus.Debug("[SERVER] Ping init")
	a.APIRouter.HandleFunc("/ping", a.getPing).Methods("GET")

	//Domain
	a.APIRouter.HandleFunc("/domains", a.getDomains).Methods("GET")
	a.APIRouter.HandleFunc("/domain", a.createDomain).Methods("POST")
	a.APIRouter.HandleFunc("/domain/{id:[0-9]+}", a.getDomain).Methods("GET")
	a.APIRouter.HandleFunc("/domain/{id:[0-9]+}", a.updateDomain).Methods("PUT")
	a.APIRouter.HandleFunc("/domain/{id:[0-9]+}", a.deleteDomain).Methods("DELETE")
	a.APIRouter.HandleFunc("/domain/{id:[0-9]+}/records", a.getDomainRecords).Methods("GET")

	//Records
	a.APIRouter.HandleFunc("/record", a.createRecord).Methods("POST")
	a.APIRouter.HandleFunc("/record/{id:[0-9]+}", a.getRecord).Methods("GET")
	a.APIRouter.HandleFunc("/record/{id:[0-9]+}", a.updateRecord).Methods("PUT")
	a.APIRouter.HandleFunc("/record/{id:[0-9]+}", a.deleteRecord).Methods("DELETE")

	//Users
	a.APIRouter.HandleFunc("/user", a.createUser).Methods("POST")
	a.APIRouter.HandleFunc("/user/{id:[0-9]+}", a.getUser).Methods("GET")
	a.APIRouter.HandleFunc("/user/self", a.getUserSelf).Methods("GET")
	a.APIRouter.HandleFunc("/user/{id:[0-9]+}", a.updateUser).Methods("PUT")
	a.APIRouter.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")
}

//Ping endpoint (to test the API)
func (a *Server) getPing(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, Response{HTTPCode: 200, Content: "Pong"})
}

//Run : Start the HTTP Server
func (a *Server) Run(conf *utils.Conf) {
	logrus.WithFields(logrus.Fields{"ip": conf.App.IP, "port": conf.App.Port}).Infof("SERVER : Started")
	addr := fmt.Sprintf("%s:%v", conf.App.IP, conf.App.Port)
	logrus.Fatal(http.ListenAndServe(addr, a.Router))
}
