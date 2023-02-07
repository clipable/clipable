package modelsx

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/volatiletech/null/v8"

	"webserver/models"
)

// UserDeserialize is a helper type for deserializng a json into a User object
type UserDeserialize jsoniter.API

// UserValidator is used to verify a User object's fields are valid
type UserValidator struct {
	*validator.Validate
}

// De/Serializer cases
var (
	UserSerialize = MakeCodec("out")

	UserDeserializeSelf UserDeserialize = MakeCodec("self-in")

	UserValidateEdit     = &UserValidator{makeValidator("validateedit")}
	UserValidateRegister = &UserValidator{makeValidator("validateregister")}
)

// User objects represent user accounts
type User struct {
	ID       string      `validateregister:"-"             validateedit:"-"                       self-in:"-"        out:"id"`
	Username null.String `validateregister:"min=2,max=64"  validateedit:"omitempty,min=2,max=64"  self-in:"username" out:"username"`
	Password null.String `validateregister:"min=2,max=256" validateedit:"omitempty,min=2,max=256" self-in:"password" out:"-"`
	JoinedAt time.Time   `validateregister:"-"             validateedit:"-"                       self-in:"-"        out:"joined_at"`
}

// ToModel converts a modelsx.User object to a model.User object
func (u *User) ToModel() *models.User {
	return &models.User{
		ID:       u.ID,
		Password: u.Password.String,
		Username: u.Username.String,
		JoinedAt: u.JoinedAt,
	}
}

// Send marshals a modelsx.User object into a sendable json byte array
func (u *User) Marshal() (int, []byte, error) {
	data, err := UserSerialize.Marshal(u)
	code := http.StatusOK

	if err != nil {
		code = http.StatusInternalServerError
	}

	return code, data, err
}

// GetUpdateWhitelist returns a list of fields from a User object that are valid
func (u *User) GetUpdateWhitelist() []string {
	nonNullFields := make([]string, 0)

	if u.Password.Valid {
		nonNullFields = append(nonNullFields, models.UserColumns.Password)
	}

	if u.Username.Valid {
		nonNullFields = append(nonNullFields, models.UserColumns.Username)
	}

	return nonNullFields
}

// UserFromModel converts a models.User object into a modelsx.User object
func UserFromModel(u *models.User) *User {
	user := &User{
		ID:       u.ID,
		Password: null.NewString(u.Password, u.Password != ""),
		Username: null.NewString(u.Username, u.Username != ""),
		JoinedAt: u.JoinedAt,
	}

	return user
}

// ParseUser parses a User object out of a client request
func ParseUser(req *http.Request, v *UserValidator) (*User, error) {
	data, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	a := &User{}

	if err := UserDeserializeSelf.Unmarshal(data, a); err != nil {
		return nil, err
	}

	if v != nil {
		if err := v.Struct(a); err != nil {
			return nil, handleValidationError(err)
		}
	}
	return a, nil
}

// UserArray is a helper type representing an array of User objects
type UserArray []*User

// Send converts a UserArray into a sendable json byte array
func (aa UserArray) Marshal() (int, []byte, error) {
	data, err := UserSerialize.Marshal(aa)
	code := http.StatusOK

	if err != nil {
		code = http.StatusInternalServerError
	}

	return code, data, err
}

// UserFromModelBatch converts multiple models.User into a modelsx.UserArray
func UserFromModelBatch(model ...*models.User) UserArray {
	var apts UserArray

	for _, m := range model {
		apts = append(apts, UserFromModel(m))
	}

	return apts
}
