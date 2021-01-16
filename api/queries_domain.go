package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/outout14/sacrebleu-api/api/types"
	"gorm.io/gorm"

	"github.com/gorilla/context"
)

//domainVerify: Verify if the domain exist and the user avec access to it
func domainVerify(err error, w http.ResponseWriter, user types.User, d types.Domain) bool {
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Domain not found.")
		return true
	}

	if !user.IsOwner(d) {
		respondWithError(w, http.StatusForbidden, "No access to this domain (no permission).")
		return true
	}
	return false
}

// getDomain endpoint.
// @Security ApiKeyAuth
// @Summary Get domain informations
// @Description Get a domain in the database by his ID
// @ID domain
// @Accept  json
// @Produce  json
// @Param   domain_id      path   int     true  "1"
// @Success 200 {object} types.Domain
// @Failure 400,403,404 {object} Response
// @Tags Domains
// @Router /domain/{domain_id} [get]
func (a *Server) getDomain(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	d := types.Domain{ID: id}

	err := d.GetDomain(a.DB)

	if domainVerify(err, w, user, d) {
		return
	}

	respondWithJSON(w, http.StatusOK, d)
}

// getDomains endpoint.
// @Security ApiKeyAuth
// @Summary Get all domains accessibles by the user
// @Description List of all domains accessibles (write & edit) according to the user permissions
// @Accept  json
// @Produce  json
// @ID domains
// @Param   count      query   int     false  "10"
// @Param   start      query   int     false  "1"
// @Success 200 {object} []types.Domain
// @Failure 400,403,404 {object} Response
// @Tags Domains
// @Router /domains [get]
func (a *Server) getDomains(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	vars := r.URL.Query()
	count, _ := strconv.Atoi(vars.Get("count"))
	start, _ := strconv.Atoi(vars.Get("start"))

	count = calcCount(count)
	start = calcStart(start)

	domains, err := types.GetDomains(a.DB, user, count, start)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, domains)
}

// getDomainRecords endpoint.
// @Security ApiKeyAuth
// @Summary Get domain records
// @Description Get domain records in the database by the domain ID
// @ID domainrecord
// @Accept  json
// @Produce  json
// @Param   domain_id      path   int     true  "1"
// @Param   count      query   int     false  "10"
// @Param   start      query   int     false  "1"
// @Success 200 {object} []types.Record
// @Failure 400,403,404 {object} Response
// @Tags Domains, Records
// @Router /domain/{domain_id}/records [get]
func (a *Server) getDomainRecords(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	//Parsing request vars
	vars := r.URL.Query()
	count, _ := strconv.Atoi(vars.Get("count"))
	start, _ := strconv.Atoi(vars.Get("start"))
	count = calcCount(count)
	start = calcStart(start)

	domainID, dbg := getID(r, w)
	if dbg {
		return
	}

	d := types.Domain{ID: domainID}
	err := d.GetDomain(a.DB)

	if domainVerify(err, w, user, d) {
		return
	}

	domains, err := d.GetDomainRecords(a.DB, count, start)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, domains)
}

// createDomain endpoint.
// @Security ApiKeyAuth
// @Summary Create domain
// @Description Create a domain in the database
// @ID newdomain
// @Accept  json
// @Produce  json
// @Success 204 {object} types.Domain
// @Failure 400,403,404,409 {object} Response
// @Tags Domains
// @Router /domain [post]
func (a *Server) createDomain(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	//Parse the submited domain
	var submitedDomain types.Domain
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&submitedDomain); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload.")
		return
	}
	defer r.Body.Close()

	//Define values
	var empty int //force "nil"
	submitedDomain.ID = empty
	if !user.IsAdmin { //Non-admin can't create domains for others
		submitedDomain.OwnerID = user.ID
	}

	if submitedDomain.Exists(a.DB) {
		respondWithError(w, http.StatusConflict, "Domain with the same FQDN already exists.")
		return
	}

	err := submitedDomain.CreateDomain(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	//Create NS records
	nameservers := a.Conf.DNS.Nameservers

	for _, nsName := range nameservers {
		nsRecord := types.Record{
			DomainID: submitedDomain.ID,
			Fqdn:     submitedDomain.Fqdn,
			Content:  nsName,
			Type:     2,
			TTL:      9600,
		}
		nsRecord.CreateRecord(a.DB)
	}

	respondWithJSON(w, http.StatusOK, submitedDomain)
}

// updateDomain endpoint.
// @Security ApiKeyAuth
// @Summary Update domain
// @Description Update a existing domain in the database by his ID (not reversible.)
// @ID putdomain
// @Accept  json
// @Produce  json
// @Param   domain_id      path   int     true  "1"
// @Success 204
// @Failure 400,403,404 {object} Response
// @Tags Domains
// @Router /domain/{domain_id} [put]
func (a *Server) updateDomain(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	d := types.Domain{ID: id}
	err := d.GetDomain(a.DB)

	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "Domain not found.")
		return
	}
	if domainVerify(err, w, user, d) {
		return
	}

	//Parse the submited domain
	submitedDomain := d
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&submitedDomain); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload.")
		return
	}
	defer r.Body.Close()

	//The domain ID and FQDN should still be the same
	submitedDomain.ID = d.ID
	submitedDomain.Fqdn = d.Fqdn

	err = submitedDomain.UpdateDomain(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	respondWithCode(w, http.StatusNoContent)
}

// deleteDomain endpoint.
// @Security ApiKeyAuth
// @Summary Delete domain
// @Description Delete a domain in the database by his ID (not reversible.)
// @ID deldomain
// @Produce  json
// @Param   domain_id      path   int     true  "1"
// @Success 204
// @Failure 400,403,404 {object} Response
// @Tags Domains
// @Router /domain/{domain_id} [delete]
func (a *Server) deleteDomain(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	d := types.Domain{ID: id}

	err := d.GetDomain(a.DB)

	if domainVerify(err, w, user, d) {
		return
	}

	//Delete all domain records
	err = d.DeleteAllDomainRecords(a.DB)
	if err != nil && err != gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusInternalServerError, "Server error.")
		return
	}

	//Delete the domain item itself
	err = d.DeleteDomain(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	respondWithCode(w, http.StatusNoContent)
}
