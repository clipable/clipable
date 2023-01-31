// Package modelsx contains developer-made types and objects (as opposed to sqlboiler-generated ones in models)
package modelsx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"github.com/volatiletech/null/v8"
)

// Unique error types
const (
	ErrorUniqueViolation  = "unique_violation"
	ErrorNotNullViolation = "not_null_violation"
)

func init() {
	jsoniter.RegisterExtension(&NullExtension{})
}

// MakeCodec creates a json de/serializer for other types
// tagKey is used to specify which de/serialize case is desired
func MakeCodec(tagKey string) jsoniter.API {
	return jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		TagKey:                 string(tagKey),
	}.Froze()
}

func nullValidator(field reflect.Value) interface{} {
	ret, _ := field.Interface().(driver.Valuer).Value()
	return ret
}

func makeValidator(tagKey string) *validator.Validate {
	ret := validator.New()
	ret.RegisterCustomTypeFunc(nullValidator, null.String{}, null.Bool{}, null.Int64{}, null.Time{})
	ret.RegisterValidation("okSymbol", stringValidator)
	ret.SetTagName(tagKey)

	return ret
}

// characterValidator returns true for ascii and extended ascii
func stringValidator(fl validator.FieldLevel) bool {
	message := fl.Field().String()
	return utf8.ValidString(message)
}

// NullExtension is a helper type enabling creating customized json encoders
type NullExtension struct {
	jsoniter.DummyExtension
}

// CreateEncoder creates a custom json encoder for handling null values
func (ne *NullExtension) CreateEncoder(t reflect2.Type) jsoniter.ValEncoder {
	if strings.HasPrefix(t.String(), "null.") {
		return &NullCodec{
			Type: t,
		}
	}

	return nil
}

// NullCodec is used for handling null values when encoding a json
type NullCodec struct {
	Type reflect2.Type
}

// NullType is a helper interface used in NullCodec methods
type NullType interface {
	IsZero() bool
	MarshalJSON() ([]byte, error)
}

// IsEmpty determines if a NullCodec field is empty or not
func (nc *NullCodec) IsEmpty(ptr unsafe.Pointer) bool {
	return nc.Type.PackEFace(ptr).(NullType).IsZero()
}

// Encode writes a NullCodec object into a stream
func (nc *NullCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if data, err := nc.Type.PackEFace(ptr).(NullType).MarshalJSON(); err == nil {
		stream.Write(data)
	}
}

// handleValidationError makes better json errors
func handleValidationError(err error) error {

	// Convert the error to a ValidationErrors
	e, ok := err.(validator.ValidationErrors)
	if !ok { // Not actually a validation error
		return err
	}

	// Create a data structure to hold the errors
	fields := map[string]string{}

	// Iterate over each individual problem field
	for i := 0; i < len(e); i++ {

		// Grab the field's name and problem
		fe, ok := e[i].(validator.FieldError)
		if !ok { // Not actually a validation error
			return err
		}
		fields[fe.Field()] = fe.Tag()
	}

	// Return a json as an error
	j, _ := json.Marshal(fields)
	return errors.New(string(j))
}
