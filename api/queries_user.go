package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/outout14/sacrebleu-api/api/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//Check if user have access to the requested user (reqUser).
func havePermissions(user types.User, reqUser types.User) bool {
	if user.IsAdmin {
		return true
	}

	if user.ID == reqUser.ID {
		return true
	}

	return false
}

// login endpoint.
// @Summary Login
// @Description User send his credentials via POST and get his token
// @ID login
// @Produce json
// @Success 200 {object} types.User
// @Failure 400,403,404 {object} Response
// @Tags Users
// @Router /login [post]
func (a *Server) login(w http.ResponseWriter, r *http.Request) {
	submitedUser := types.User{}
	submitedUser.Username = r.FormValue("username")
	submitedUser.Password = r.FormValue("password")

	resultUser := submitedUser

	err := resultUser.GetUserByUsername(a.DB)

	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusForbidden, "Credentials don't match.")
		return
	}

	//Check for the password
	err = bcrypt.CompareHashAndPassword([]byte(resultUser.Password), []byte(submitedUser.Password))
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Credentials don't match.")
	} else {
		respondWithJSON(w, http.StatusOK, resultUser)
	}
}

// getUser endpoint.
// @Security ApiKeyAuth
// @Summary Get user informations
// @Description Get a user in the database by his ID
// @ID user
// @Produce  json
// @Param   user_id      path   int     true  "1"
// @Success 200 {object} types.User
// @Failure 400,403,404 {object} Response
// @Tags Users
// @Router /user/{user_id} [get]
func (a *Server) getUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	u := types.User{ID: id}

	err := u.GetUser(a.DB)
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "User not found.")
		return
	}
	if checkSrvErr(err, w) {
		return
	}

	if !havePermissions(user, u) {
		respondWithError(w, http.StatusForbidden, "No access to this user (no permission).")
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

// getUserSelf endpoint.
// @Security ApiKeyAuth
// @Summary Get logged user informations
// @Description Get the user object of who is running the query.
// @ID userself
// @Produce  json
// @Success 200 {object} types.User
// @Tags Users
// @Router /user/self [get]
func (a *Server) getUserSelf(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user
	respondWithJSON(w, http.StatusOK, user)
}

// createUser endpoint.
// @Security ApiKeyAuth
// @Summary Create user
// @Description Create a user in the database
// @ID newuser
// @Accept  json
// @Produce  json
// @Success 204 {object} types.User
// @Failure 400,403,404,409 {object} Response
// @Tags Users
// @Router /user [post]
func (a *Server) createUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	if !user.IsAdmin { //Non-admin can't create users !
		respondWithError(w, http.StatusForbidden, "No access to this functionality (no permission).")
		return
	}

	//Parse the submited user
	var submitedUser types.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&submitedUser); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload.")
		return
	}
	defer r.Body.Close()

	//Define values
	var empty int //force "nil"
	submitedUser.ID = empty
	submitedUser.Token = GenerateToken(submitedUser.Email)
	submitedUser.Password, _ = HashPassword(submitedUser.Password)

	err := submitedUser.CreateUser(a.DB)
	if err == gorm.ErrRegistered {
		respondWithError(w, http.StatusConflict, "User with the same email or username already exists.")
		return
	}
	if checkSrvErr(err, w) {
		return
	}

	respondWithJSON(w, http.StatusOK, submitedUser)
}

// updateUser endpoint.
// @Security ApiKeyAuth
// @Summary Update user informations
// @Description Update a user in the database by his ID (not reversible.)
// @ID user
// @Accept  json
// @Produce  json
// @Param   user_id      path   int     true  "1"
// @Success 200 {object} types.User
// @Failure 400,403,404 {object} Response
// @Tags Users
// @Router /user/{user_id} [put]
func (a *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	u := types.User{ID: id}

	err := u.GetUser(a.DB)
	if !havePermissions(user, u) {
		respondWithError(w, http.StatusForbidden, "No access to this user (no permission).")
		return
	}
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "User not found.")
		return
	}

	//Parse the submited user
	submitedUser := u
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&submitedUser); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload.")
		return
	}
	defer r.Body.Close()

	//Check if modified username or email already exist
	if submitedUser.Username != u.Username {
		if submitedUser.UsernameExists(a.DB) {
			respondWithError(w, http.StatusConflict, "User with the same username already exists.")
			return
		}
	}
	if submitedUser.Email != u.Email {
		if submitedUser.EmailExists(a.DB) {
			logrus.Debug("OK")
			respondWithError(w, http.StatusConflict, "User with the same email already exists.")
			return
		}
	}

	//The user ID & token should still be the same
	submitedUser.ID = u.ID
	submitedUser.Token = u.Token
	if !user.IsAdmin {
		submitedUser.IsAdmin = false
	}
	submitedUser.Password, _ = HashPassword(submitedUser.Password)

	err = submitedUser.UpdateUser(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	respondWithJSON(w, http.StatusOK, submitedUser)
}

// deleteUser endpoint.
// @Security ApiKeyAuth
// @Summary Delete user
// @Description Delete a user in the database by his ID (not reversible.)
// @ID user
// @Produce  json
// @Param   user_id      path   int     true  "1"
// @Success 204
// @Failure 400,403,404 {object} Response
// @Tags Users
// @Router /user/{user_id} [delete]
func (a *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	u := types.User{ID: id}

	err := u.GetUser(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	if !havePermissions(user, u) {
		respondWithError(w, http.StatusForbidden, "No access to this user (no permission).")
		return
	}

	err = u.DeleteUser(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	respondWithCode(w, http.StatusNoContent)
}
