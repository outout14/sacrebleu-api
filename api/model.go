package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

//App : Struct for App (http server) configuration in the config.ini file
type App struct {
	Router  *mux.Router
	DB      *gorm.DB
	Port    int
	IP      string
	Logdir  string
	Logfile bool
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) getDemo(w http.ResponseWriter, r *http.Request) {

	respondWithJSON(w, http.StatusOK, "wow")
}
