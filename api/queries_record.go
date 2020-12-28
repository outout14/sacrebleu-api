package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/outout14/sacrebleu-api/api/types"
	"gorm.io/gorm"
)

// getRecord endpoint.
// @Security ApiKeyAuth
// @Summary Get record informations
// @Description Get a record in the database by his ID
// @ID record
// @Accept  json
// @Produce  json
// @Param   record_id      path   int     true  "1"
// @Success 200 {object} types.Record
// @Failure 400,403,404 {object} Response
// @Tags Records
// @Router /record/{record_id} [get]
func (a *Server) getRecord(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	record := types.Record{ID: id}
	err := record.GetRecord(a.DB)
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "Record not found.")
		return
	}

	//Check parent domain permissions
	parentDomain := types.Domain{ID: record.DomainID}
	err = parentDomain.GetOwner(a.DB)
	if domainVerify(err, w, user, parentDomain) {
		return
	}

	respondWithJSON(w, http.StatusOK, record)
}

// createRecord endpoint.
// @Security ApiKeyAuth
// @Summary Create record
// @Description Create a record in the database
// @ID newrecord
// @Accept  json
// @Produce  json
// @Success 204 {object} types.Record
// @Failure 400,403,404,409 {object} Response
// @Tags Records
// @Router /record [post]
func (a *Server) createRecord(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	//Parse the submited record
	var submitedRecord types.Record
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&submitedRecord); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload.")
		return
	}
	defer r.Body.Close()

	//Check parent domain permissions
	parentDomain := types.Domain{ID: submitedRecord.DomainID}
	err := parentDomain.GetDomain(a.DB)
	if domainVerify(err, w, user, parentDomain) {
		return
	}

	//Define values
	var empty int //force "nil"
	submitedRecord.ID = empty

	//Check if record is added to the correct domain
	if !strings.HasSuffix(submitedRecord.Fqdn, parentDomain.Fqdn) {
		respondWithError(w, http.StatusBadRequest, "Record FQDN end don't correspond to parent domain FQDN.")
		return
	}

	err = submitedRecord.CreateRecord(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	parentDomain.UpdateSOA(a.DB, user)

	respondWithJSON(w, http.StatusOK, submitedRecord)
}

// updateRecord endpoint.
// @Security ApiKeyAuth
// @Summary Update record
// @Description Update a existing record in the database by his ID (not reversible.)
// @ID putrecord
// @Accept  json
// @Produce  json
// @Param   record_id      path   int     true  "1"
// @Success 204
// @Failure 400,403,404 {object} Response
// @Tags Records
// @Router /record/{record_id} [put]
func (a *Server) updateRecord(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	//Get actual record
	record := types.Record{ID: id}
	err := record.GetRecord(a.DB)
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "Record not found.")
		return
	}

	//Check parent domain permissions
	d := types.Domain{ID: record.DomainID}
	err = d.GetDomain(a.DB)
	if domainVerify(err, w, user, d) {
		return
	}

	//Parse the submited domain
	submitedRecord := record
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&submitedRecord); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload.")
		return
	}
	defer r.Body.Close()

	//The record ID and Domain ID should still be the same
	submitedRecord.ID = record.ID
	submitedRecord.DomainID = record.DomainID

	err = submitedRecord.UpdateRecord(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	d.UpdateSOA(a.DB, user)

	respondWithCode(w, http.StatusNoContent)
}

// deleteRecord endpoint.
// @Security ApiKeyAuth
// @Summary Delete record
// @Description Delete a record in the database by his ID (not reversible.)
// @ID delrecord
// @Produce  json
// @Param   record_id      path   int     true  "1"
// @Success 204
// @Failure 400,403,404 {object} Response
// @Tags Records
// @Router /record/{record_id} [delete]
func (a *Server) deleteRecord(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(types.User) //avoid asking the SQL server again for the user

	id, dbg := getID(r, w)
	if dbg {
		return
	}

	//Get actual record
	record := types.Record{ID: id}
	err := record.GetRecord(a.DB)
	//Check parent domain permissions
	d := types.Domain{ID: record.DomainID}
	err = d.GetDomain(a.DB)
	if domainVerify(err, w, user, d) {
		return
	}

	err = record.DeleteRecord(a.DB)
	if checkSrvErr(err, w) {
		return
	}

	d.UpdateSOA(a.DB, user)

	respondWithCode(w, http.StatusNoContent)
}
