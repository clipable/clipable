package routes

import (
	"database/sql"
	"net/http"
	"webserver/modelsx"

	"github.com/alexedwards/argon2id"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const SESSION_NAME = "webserver"
const SESSION_KEY_UID = "uid"
const SESSION_KEY_OAUTH_STATE = "oauth-state"
const SESSION_KEY_ID = "id"

func (r *Routes) AllowRegistration(resp http.ResponseWriter, req *http.Request) (int, []byte, error) {
	if !r.cfg.AllowRegistration {
		return http.StatusForbidden, nil, nil
	}

	return http.StatusOK, nil, nil
}

func (r *Routes) Login(resp http.ResponseWriter, req *http.Request) (int, []byte, error) {
	session, _ := r.store.Get(req, SESSION_NAME)

	json, err := modelsx.ParseUser(req, modelsx.UserValidateRegister)

	if err != nil {
		return http.StatusBadRequest, []byte(err.Error()), nil
	}

	user, err := r.Users.FindUsername(req.Context(), json.Username.String)

	if err == sql.ErrNoRows {
		return http.StatusUnauthorized, []byte("Invalid username/password combination"), nil
	} else if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to find user")
	}

	match, err := argon2id.ComparePasswordAndHash(json.Password.String, user.Password)

	if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to compare password")
	}

	if !match {
		return http.StatusUnauthorized, []byte("Invalid username/password combination"), nil
	}

	session.Values[SESSION_KEY_ID] = user.ID
	if err := session.Save(req, resp); err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to save session")
	}

	return modelsx.UserFromModel(user).Marshal()
}

func (r *Routes) Logout(resp http.ResponseWriter, req *http.Request) (int, []byte, error) {
	session, _ := r.store.Get(req, SESSION_NAME)

	session.Options.MaxAge = -1

	if err := session.Save(req, resp); err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to save session")
	}

	resp.Header().Set("Clear-Site-Data", `"cookies"`)

	return http.StatusOK, nil, nil
}

func (r *Routes) Register(resp http.ResponseWriter, req *http.Request) (int, []byte, error) {
	if !r.cfg.AllowRegistration {
		return http.StatusForbidden, []byte("Registration is disabled"), nil
	}

	usr, err := modelsx.ParseUser(req, modelsx.UserValidateRegister)

	if err != nil {
		return http.StatusBadRequest, []byte(err.Error()), nil
	}

	exists, err := r.Users.ExistsUsername(req.Context(), usr.Username.String)

	if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to check if username exists")
	}

	if exists {
		return http.StatusConflict, []byte("Username already exists"), nil
	}

	hash, err := argon2id.CreateHash(usr.Password.String, argon2id.DefaultParams)

	if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to hash password")
	}

	usr.Password.String = hash

	model := usr.ToModel()

	if err := r.Users.Create(req.Context(), model, boil.Whitelist(usr.GetUpdateWhitelist()...)); err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to create user")
	}

	session, _ := r.store.Get(req, SESSION_NAME)

	session.Values[SESSION_KEY_ID] = model.ID
	if err := session.Save(req, resp); err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "Failed to save session")
	}

	return modelsx.UserFromModel(model).Marshal()
}
