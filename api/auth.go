package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/outout14/sacrebleu-api/api/types"

	"github.com/sirupsen/logrus"
)

//JwtVerify : Token verification
func JwtVerify(a *Server) func(http.Handler) http.Handler {
	//Method to pass the *Server struct and get access to the SQL DB
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var header = r.Header.Get("x-access-token")

			json.NewEncoder(w).Encode(r)
			header = strings.TrimSpace(header)

			if header == "" {
				logrus.Debug("AUTH : Token header not found.")
				respondWithError(w, http.StatusForbidden, "Missing access token.")
				return
			}

			var user types.User
			result := a.DB.Where("token = ?", header).First(&user)
			if result.Error != nil {
				logrus.WithFields(logrus.Fields{"x-access-token": header}).Debug("AUTH : Token invalid.")
				respondWithError(w, http.StatusForbidden, "Token invalid.")
				return
			}

			//Will be passed to the request function to avoid asking the SQL server again for the user
			context.Set(r, "user", user)
			next.ServeHTTP(w, r)
		})
	}
}
