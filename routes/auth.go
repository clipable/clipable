package routes

import (
	"net/http"
	"webserver/models"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const SESSION_NAME = "webserver"
const SESSION_KEY_UID = "uid"
const SESSION_KEY_OAUTH_STATE = "oauth-state"
const SESSION_KEY_ID = "id"

func (r *Routes) Login(resp http.ResponseWriter, req *http.Request) {
	session, _ := r.store.Get(req, SESSION_NAME)

	state := uuid.New().String()

	session.Values[SESSION_KEY_OAUTH_STATE] = state
	session.Save(req, resp)

	http.Redirect(resp, req, r.oauth.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (r *Routes) Logout(resp http.ResponseWriter, req *http.Request) {
	session, _ := r.store.Get(req, SESSION_NAME)

	session.Options.MaxAge = -1
	session.Save(req, resp)
	resp.Header().Set("Clear-Site-Data", `"cookies", "storage"`)
}

func (r *Routes) Callback(resp http.ResponseWriter, req *http.Request) {
	session, _ := r.store.Get(req, SESSION_NAME)

	state, ok := session.Values[SESSION_KEY_OAUTH_STATE]

	if !ok {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.FormValue("state") != state.(string) {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	oat, err := r.oauth.Exchange(req.Context(), req.FormValue("code"))

	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		log.Errorln(err)
		return
	}

	parser := &jwt.Parser{}
	jwt, _, err := parser.ParseUnverified(oat.AccessToken, &ID{})

	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		log.Errorln(err)
		return
	}

	token := jwt.Claims.(*ID)

	user := &models.User{
		ID:        token.UUID,
		Firstname: token.GivenName,
		Lastname:  token.FamilyName,
		Email:     token.Email,
	}

	if ok, err := r.Users.Exists(req.Context(), token.UUID); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		log.Errorln(err)
		return
	} else if !ok { // If the user doesn't exist, create them
		if err := r.Users.Create(req.Context(), user, boil.Whitelist(models.UserColumns.ID, models.UserColumns.Firstname, models.UserColumns.Lastname, models.UserColumns.Email)); err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			log.Errorln(err)
			return
		}
	} else if ok { // If the user does exist, update their fields to re-sync their profile
		if err := r.Users.Update(req.Context(), user, boil.Whitelist(models.UserColumns.ID, models.UserColumns.Firstname, models.UserColumns.Lastname, models.UserColumns.Email)); err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			log.Errorln(err)
			return
		}
	}

	session.Values[SESSION_KEY_ID] = token

	session.Save(req, resp)

	http.Redirect(resp, req, r.cfg.CORS.Origin+"/", http.StatusTemporaryRedirect)
}

type ID struct {
	jwt.StandardClaims
	UserName      string   `json:"user_name"`
	TokenType     string   `json:"token_type"`
	UUID          string   `json:"uuid"`
	ClientID      string   `json:"client_id"`
	Sid           string   `json:"sid"`
	GroupStrs     []string `json:"group_strs"`
	UpdatedAt     int      `json:"updated_at"`
	GrantType     string   `json:"grant_type"`
	Azp           string   `json:"azp"`
	AuthTime      int      `json:"auth_time"`
	Scope         string   `json:"scope"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	GivenName     string   `json:"given_name"`
	MiddleName    string   `json:"middle_name"`
	Nonce         string   `json:"nonce"`
	Authorities   []string `json:"authorities"`
	Name          string   `json:"name"`
	FamilyName    string   `json:"family_name"`
	Username      string   `json:"username"`
}
