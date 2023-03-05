package routes

import (
	"database/sql"
	"net/http"
	"webserver/models"
	"webserver/modelsx"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// UpdateUser updates a user
//
// Success when provided valid User json; returns updated User json
// Normal users can only update themselves
//
// PATCH /users/{user id}
func (r *Routes) UpdateUser(user *models.User, req *http.Request) (int, []byte, error) {
	if user == nil {
		return http.StatusUnauthorized, nil, nil
	}

	vars := vars(req)

	// Users shouldn't be able to update another users
	if vars.UID != user.ID {
		return http.StatusForbidden, nil, nil
	}

	updateUser, err := modelsx.ParseUser(req, modelsx.UserValidateEdit)

	if err != nil {
		return http.StatusBadRequest, []byte(err.Error()), nil
	}

	updateUser.ID = modelsx.HashID(vars.UID)

	model := updateUser.ToModel()

	if err := r.Users.Update(req.Context(), model, boil.Whitelist(updateUser.GetUpdateWhitelist()...)); err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "failed to update user")
	}

	return modelsx.UserFromModel(model).Marshal()
}

// SearchUsers returns list of users available to user matching query string
//
// Success returns json array of User objects representing users the requesting User can see, and which match query string
// Non-admins cannot see banned users
//
// GET /users/search
func (r *Routes) SearchUsers(user *models.User, req *http.Request) (int, []byte, error) {
	users, err := r.Users.SearchMany(req.Context(), req.URL.Query().Get("query"))

	if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "failed to search users")
	}

	if len(users) == 0 {
		return http.StatusNoContent, nil, nil
	}

	return modelsx.UserFromModelBatch(users...).Marshal()
}

// GetUser returns specified user
//
// # Success returns json of User with spcified id
//
// GET /users/{user id}
func (r *Routes) GetUser(user *models.User, req *http.Request) (int, []byte, error) {
	vars := vars(req)

	targetUser, err := r.Users.Find(req.Context(), vars.UID)

	if err == sql.ErrNoRows {
		return http.StatusNotFound, nil, nil
	} else if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "failed to find user")
	}

	return modelsx.UserFromModel(targetUser).Marshal()
}

// GetCurrentUser returns json of requesting Uset
//
// # Success returns json of User making request
//
// GET /users/me
func (r *Routes) GetCurrentUser(user *models.User, req *http.Request) (int, []byte, error) {
	if user == nil {
		return http.StatusUnauthorized, nil, nil
	}
	return modelsx.UserFromModel(user).Marshal()
}

// GetCurrentUserClips returns list of clips created by requesting user
//
// # Success returns json array of Clip objects created by requesting User
//
// GET /users/me/clips
func (r *Routes) GetUsersClips(user *models.User, req *http.Request) (int, []byte, error) {
	vars := vars(req)

	clips, err := r.Clips.FindMany(req.Context(), user, modelsx.NewBuilder().
		Add(models.ClipWhere.CreatorID.EQ(vars.UID)).
		Add(getPaginationMods(req, models.ClipColumns.CreatedAt, models.TableNames.Clips, models.ClipColumns.ID)...,
		)...,
	)

	if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "failed to find clips")
	}

	if len(clips) == 0 {
		return http.StatusNoContent, nil, nil
	}

	return modelsx.ClipFromModelBatch(clips...).Marshal()
}

// GetUsers returns list of all users available to user
//
// # Success returns json array of all User objects available to requesting User
//
// GET /users
func (r *Routes) GetUsers(user *models.User, req *http.Request) (int, []byte, error) {
	users, err := r.Users.FindMany(req.Context(), getPaginationMods(req, models.UserColumns.JoinedAt, models.TableNames.User, models.UserColumns.ID)...)

	if err != nil {
		return http.StatusInternalServerError, nil, errors.Wrap(err, "failed to find users")
	}

	if len(users) == 0 {
		return http.StatusNoContent, nil, nil
	}

	return modelsx.UserFromModelBatch(users...).Marshal()
}
