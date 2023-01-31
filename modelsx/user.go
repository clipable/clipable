package modelsx

import (
	"io/ioutil"
	"net/http"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/volatiletech/null/v8"

	"webserver/models"
)

// UserSerialize is a helper type for serializing a User into a json byte array
type UserSerialize jsoniter.API

// UserDeserialize is a helper type for deserializng a json into a User object
type UserDeserialize jsoniter.API

// UserValidator is used to verify a User object's fields are valid
type UserValidator struct {
	*validator.Validate
}

// De/Serializer cases
var (
	UserSerializeSelf UserSerialize = MakeCodec("self-out")
	UserSerializeUser UserSerialize = MakeCodec("user-out")

	UserDeserializeSelf UserDeserialize = MakeCodec("self-in")

	UserValidateEdit = &UserValidator{makeValidator("validateedit")}
)

// User objects represent user accounts
type User struct {
	ID        string      `validateedit:"-"                      self-in:"-"            self-out:"id"                 user-out:"id"`
	Username  null.String `validateedit:"omitempty,min=2,max=64" self-in:"username"     self-out:"username,omitempty" user-out:"username,omitempty"`
	Email     string      `validateedit:"-"                      self-in:"-"            self-out:"email,omitempty"    user-out:"email,omitempty"`
	Firstname string      `validateedit:"-"                      self-in:"-"            self-out:"firstName"          user-out:"firstName"`
	Lastname  string      `validateedit:"-"                      self-in:"-"            self-out:"lastName"           user-out:"lastName"`
	Online    bool        `validateedit:"-"                      self-in:"online"       self-out:"online"             user-out:"online"`
	Phone     null.String `validateedit:"-"                      self-in:"phone"        self-out:"phone,omitempty"    user-out:"phone,omitempty"`
}

// ToModel converts a modelsx.User object to a model.User object
func (u *User) ToModel() *models.User {
	return &models.User{
		ID:        u.ID,
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Username:  u.Username.String,
		Online:    u.Online,
		Phone:     u.Phone.String,
	}
}

// Send marshals a modelsx.User object into a sendable json byte array
func (u *User) Marshal(codec UserSerialize) (int, []byte, error) {
	data, err := codec.Marshal(u)
	code := http.StatusOK

	if err != nil {
		code = http.StatusInternalServerError
	}

	return code, data, err
}

// GetUpdateWhitelist returns a list of fields from a User object that are valid
func (u *User) GetUpdateWhitelist() []string {
	nonNullFields := make([]string, 0)

	if u.Email != "" {
		nonNullFields = append(nonNullFields, models.UserColumns.Email)
	}

	if u.Firstname != "" {
		nonNullFields = append(nonNullFields, models.UserColumns.Firstname)
	}

	if u.Lastname != "" {
		nonNullFields = append(nonNullFields, models.UserColumns.Lastname)
	}

	if u.Username.Valid {
		nonNullFields = append(nonNullFields, models.UserColumns.Username)
	}

	if u.Phone.Valid {
		nonNullFields = append(nonNullFields, models.UserColumns.Phone)
	}

	return nonNullFields
}

// UserFromModel converts a models.User object into a modelsx.User object
func UserFromModel(u *models.User) *User {
	user := &User{
		ID:        u.ID,
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Username:  null.StringFrom(u.Username),
		Online:    u.Online,
		Phone:     null.StringFrom(u.Phone),
	}

	return user
}

// ParseUser parses a User object out of a client request
func ParseUser(req *http.Request, codec UserDeserialize, v *UserValidator) (*User, error) {
	data, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	a := &User{}

	if err := codec.Unmarshal(data, a); err != nil {
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
func (aa UserArray) Marshal(codec UserSerialize) (int, []byte, error) {
	data, err := codec.Marshal(aa)
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
