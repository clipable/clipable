package modelsx

import (
	"io"
	"net/http"
	"time"

	"webserver/models"

	. "github.com/docker/go-units"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
)

// De/Serializer cases
var (
	ClipSerialize = MakeCodec("out")

	ClipDeserialize = MakeCodec("in")

	ClipValidate = makeValidator("validate")
)

// Clip objects represent Clip accounts
type Clip struct {
	ID          HashID      `validate:"-"                  in:"-"           out:"id"                   `
	Title       string      `validate:"min=2,max=64"       in:"title"       out:"title"                `
	Description null.String `validate:"omitempty,max=1024" in:"description" out:"description,omitempty"`
	CreatedAt   time.Time   `validate:"-"                  in:"-"           out:"created_at"           `
	CreatorID   HashID      `validate:"-"                  in:"-"           out:"-"                    `
	Processing  bool        `validate:"-"                  in:"-"           out:"processing"           `
	Unlisted    null.Bool   `validate:"-"                  in:"unlisted"    out:"unlisted"             `
	Views       int64       `validate:"-"                  in:"-"           out:"views"                `

	Creator *User `validate:"-" in:"-" out:"creator"`
}

// ToModel converts a modelsx.Clip object to a model.Clip object
func (u *Clip) ToModel() *models.Clip {
	return &models.Clip{
		ID:          int64(u.ID),
		Title:       u.Title,
		Description: u.Description,
		CreatedAt:   u.CreatedAt,
		CreatorID:   int64(u.CreatorID),
		Processing:  u.Processing,
		Unlisted:    u.Unlisted.Bool,
		Views:       u.Views,
	}
}

// Send marshals a modelsx.Clip object into a sendable json byte array
func (u *Clip) Marshal() (int, []byte, error) {
	data, err := ClipSerialize.Marshal(u)
	code := http.StatusOK

	if err != nil {
		code = http.StatusInternalServerError
	}

	return code, data, err
}

// GetUpdateWhitelist returns a list of fields from a Clip object that are valid
func (u *Clip) GetUpdateWhitelist() []string {
	nonNullFields := make([]string, 0)

	if u.Title != "" {
		nonNullFields = append(nonNullFields, models.ClipColumns.Title)
	}

	if u.Description.Valid {
		nonNullFields = append(nonNullFields, models.ClipColumns.Description)
	}

	if u.CreatorID != 0 {
		nonNullFields = append(nonNullFields, models.ClipColumns.CreatorID)
	}

	if u.Unlisted.Valid {
		nonNullFields = append(nonNullFields, models.ClipColumns.Unlisted)
	}

	return nonNullFields
}

// ClipFromModel converts a models.Clip object into a modelsx.Clip object
func ClipFromModel(u *models.Clip) *Clip {
	Clip := &Clip{
		ID:          HashID(u.ID),
		Title:       u.Title,
		Description: u.Description,
		CreatedAt:   u.CreatedAt,
		CreatorID:   HashID(u.CreatorID),
		Processing:  u.Processing,
		Unlisted:    null.BoolFrom(u.Unlisted),
		Views:       u.Views,
	}

	if u.R != nil {
		Clip.Creator = UserFromModel(u.R.Creator)
	}

	return Clip
}

// ParseClip parses a Clip object out of a client request
func ParseClip(req io.Reader) (*Clip, error) {
	data, err := io.ReadAll(io.LimitReader(req, 2*KB))

	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}

	a := &Clip{}

	if err := ClipDeserialize.Unmarshal(data, a); err != nil {
		return nil, errors.Wrap(err, "failed to parse request body")
	}

	if err := ClipValidate.Struct(a); err != nil {
		return nil, handleValidationError(err)
	}

	return a, nil
}

// ClipArray is a helper type representing an array of Clip objects
type ClipArray []*Clip

// Send converts a ClipArray into a sendable json byte array
func (aa ClipArray) Marshal() (int, []byte, error) {
	data, err := ClipSerialize.Marshal(aa)
	code := http.StatusOK

	if err != nil {
		code = http.StatusInternalServerError
	}

	return code, data, err
}

// ClipFromModelBatch converts multiple models.Clip into a modelsx.ClipArray
func ClipFromModelBatch(model ...*models.Clip) ClipArray {
	var apts ClipArray

	for _, m := range model {
		apts = append(apts, ClipFromModel(m))
	}

	return apts
}
