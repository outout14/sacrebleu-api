package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/outout14/sacrebleu-dns/utils"
	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

//Server : Struct for App (http server) configuration in the config.ini file
type Server struct {
	Router    *mux.Router
	APIRouter *mux.Router
	DB        *gorm.DB
	Conf      *utils.Conf
}

//Response : Used to reply to http query
type Response struct {
	HTTPCode int    `example:"403"`
	Content  string `example:"Token invalid."`
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, Response{HTTPCode: code, Content: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithCode(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func checkSrvErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error.")
		return true
	}
	return false
}

func calcStart(start int) int {
	if start < 0 {
		start = 0
	}
	return start
}

func calcCount(count int) int {
	if count > 10 || count < 1 {
		count = 10
	}
	return count
}

//getID: get the GET id parameter
func getID(r *http.Request, w http.ResponseWriter) (int, bool) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid query ID")
		return 0, true
	}
	return id, false
}

//HashPassword : just bcrypt stuff.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

//GenerateToken : generate user token based on the seed (eg : email)
func GenerateToken(seed string) string {
	b := make([]byte, 8)
	rand.Read(b)
	token := fmt.Sprintf("%s%x", seed, b)
	return base64.StdEncoding.EncodeToString([]byte(token))
}
