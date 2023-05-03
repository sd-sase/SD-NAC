package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/inverse-inc/go-utils/log"
	"github.com/inverse-inc/packetfence/go/db"
	"github.com/inverse-inc/packetfence/go/plugin/caddy2/dal/models"
	"github.com/inverse-inc/packetfence/go/plugin/caddy2/pfpki/sql"
	"github.com/inverse-inc/packetfence/go/plugin/caddy2/pfpki/types"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
)

type DnsAuditLog struct {
	DB  *gorm.DB
	Ctx *context.Context
}

func NewDnsAuditLog() *DnsAuditLog {
	DB, err := gorm.Open("mysql", db.ReturnURIFromConfig(context.Background()))
	ctx := context.Background()
	if err != nil {
		log.LoggerWContext(ctx).Warn(err.Error())
	}
	return &DnsAuditLog{
		DB:  DB,
		Ctx: &ctx,
	}
}

func (a *DnsAuditLog) List(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	model := models.NewDnsAuditLogModel(a.DB, a.Ctx)
	var body RespBody
	var err error
	body.Status = http.StatusOK

	vars, err := types.DecodeUrlQuery(r)
	if err != nil {
		setError(&body, err, http.StatusBadRequest)
		outputResult(w, body)
		return
	}

	body.DBRes, err = model.Paginated(vars)
	if err != nil {
		setError(&body, err, http.StatusInternalServerError)
		outputResult(w, body)
		return
	}

	outputResult(w, body)
}

func (a *DnsAuditLog) Search(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	model := models.NewDnsAuditLogModel(a.DB, a.Ctx)
	var body RespBody
	var err error
	body.Status = http.StatusOK

	var vars sql.Vars
	err = vars.DecodeBodyJson(r)
	if err != nil {
		setError(&body, err, http.StatusBadRequest)
		outputResult(w, body)
		return
	}

	body.DBRes, err = model.Search(vars)
	if err != nil {
		setError(&body, err, http.StatusNotFound)
		outputResult(w, body)
		return
	}
	outputResult(w, body)
}

func (a *DnsAuditLog) GetItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	model := models.NewDnsAuditLogModel(a.DB, a.Ctx)
	var body RespBody
	var err error
	body.Status = http.StatusOK

	id := p.ByName("id")
	_, err = strconv.Atoi(id)
	if err != nil {
		setError(&body, errors.New("invalid format for dns audit log entry ID"), http.StatusBadRequest)
		outputResult(w, body)
		return
	}

	body.DBRes, err = model.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			setError(&body, err, http.StatusNotFound)
		} else {
			setError(&body, err, http.StatusInternalServerError)
		}
		outputResult(w, body)
		return
	}
	outputResult(w, body)
}

func (a *DnsAuditLog) DeleteItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	model := models.NewDnsAuditLogModel(a.DB, a.Ctx)
	var body RespBody
	var err error
	body.Status = http.StatusOK

	id := p.ByName("id")
	_, err = strconv.Atoi(id)
	if err != nil {
		setError(&body, errors.New("invalid format for dns audit log entry ID"), http.StatusBadRequest)
		outputResult(w, body)
		return
	}

	body.DBRes, err = model.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			setError(&body, err, http.StatusNotFound)
		} else {
			setError(&body, err, http.StatusInternalServerError)
		}
		outputResult(w, body)
		return
	}

	body.Message = fmt.Sprintf("Deleted %s successfully", id)
	outputResult(w, body)
}

func (a *DnsAuditLog) UpdateItem(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	model := models.NewDnsAuditLogModel(a.DB, a.Ctx)
	var body RespBody
	var err error
	body.Status = http.StatusOK

	id := p.ByName("id")
	nID, err := strconv.Atoi(id)
	if err != nil {
		setError(&body, errors.New("invalid format for dns audit log entry ID"), http.StatusBadRequest)
		outputResult(w, body)
		return
	}

	body.DBRes, err = model.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			setError(&body, err, http.StatusNotFound)
		} else {
			setError(&body, err, http.StatusInternalServerError)
		}
		outputResult(w, body)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		setError(&body, err, http.StatusBadRequest)
		outputResult(w, body)
		return
	}
	err = json.Unmarshal(payload, &model)
	model.ID = int64(nID)
	if err != nil {
		setError(&body, err, http.StatusUnprocessableEntity)
		outputResult(w, body)
		return
	}

	body.DBRes, err = model.Update()

	if err != nil {
		setError(&body, err, http.StatusUnprocessableEntity)
		outputResult(w, body)
		return
	}
	body.Message = fmt.Sprintf("id %d updated", nID)
	outputResult(w, body)
}

func (a *DnsAuditLog) AddToRouter(r *httprouter.Router) {
	r.GET("/api/v1/dns_audit_logs", a.List)
	r.POST("/api/v1/dns_audit_logs/search", a.Search)
	r.GET("/api/v1/dns_audit_log/:id", a.GetItem)
}
