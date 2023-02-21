package routes

import (
	"database/sql"
	"net/http"
	"webserver/modelsx"

	"github.com/alexedwards/argon2id"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const SESSION_NAME = "webserver"
const SESSION_KEY_UID = "uid"
const SESSION_KEY_OAUTH_STATE = "oauth-state"
const SESSION_KEY_ID = "id"

func (r *Routes) Login(resp http.ResponseWriter, req *http.Request) {
	session, _ := r.store.Get(req, SESSION_NAME)

	json, err := modelsx.ParseUser(req, modelsx.UserValidateRegister)

	if err != nil {
		r.handleErr(resp, http.StatusBadRequest, err, "Failed to parse user")
		return
	}

	user, err := r.Users.FindUsername(req.Context(), json.Username.String)

	if err == sql.ErrNoRows {
		r.handleErr(resp, http.StatusUnauthorized, err, "Invalid username or password")
		return
	} else if err != nil {
		r.handleErr(resp, http.StatusInternalServerError, err, "Failed to find user")
		return
	}

	match, err := argon2id.ComparePasswordAndHash(json.Password.String, user.Password)

	if err != nil {
		r.handleErr(resp, http.StatusInternalServerError, err, "Failed to compare password")
		return
	}

	if !match {
		r.handleErr(resp, http.StatusUnauthorized, nil, "Invalid username or password")
		return
	}

	session.Values[SESSION_KEY_ID] = user.ID
	session.Save(req, resp)

	resp.WriteHeader(http.StatusOK)
}

func (r *Routes) Logout(resp http.ResponseWriter, req *http.Request) {
	session, _ := r.store.Get(req, SESSION_NAME)

	session.Options.MaxAge = -1
	session.Save(req, resp)
	resp.Header().Set("Clear-Site-Data", `"cookies", "storage"`)
}

func (r *Routes) Register(resp http.ResponseWriter, req *http.Request) {

	usr, err := modelsx.ParseUser(req, modelsx.UserValidateRegister)

	if err != nil {
		r.handleErr(resp, http.StatusBadRequest, err, "Failed to parse user")
		return
	}

	exists, err := r.Users.ExistsUsername(req.Context(), usr.Username.String)

	if err != nil {
		r.handleErr(resp, http.StatusInternalServerError, err, "Failed to check if user exists")
		return
	}

	if exists {
		r.handleErr(resp, http.StatusConflict, nil, "Username already exists")
		return
	}

	hash, err := argon2id.CreateHash(usr.Password.String, argon2id.DefaultParams)

	if err != nil {
		r.handleErr(resp, http.StatusInternalServerError, nil, "Failed to hash password")
		return
	}

	usr.Password.String = hash

	model := usr.ToModel()

	if err := r.Users.Create(req.Context(), model, boil.Whitelist(usr.GetUpdateWhitelist()...)); err != nil {
		r.handleErr(resp, http.StatusInternalServerError, nil, "Failed to create user")
		return
	}

	session, _ := r.store.Get(req, SESSION_NAME)

	session.Values[SESSION_KEY_ID] = model.ID
	session.Save(req, resp)

	res, body, err := modelsx.UserFromModel(model).Marshal()

	if err != nil {
		r.handleErr(resp, http.StatusInternalServerError, nil, "Failed to marshal user")
		return
	}

	resp.WriteHeader(res)
	resp.Write(body)
}
