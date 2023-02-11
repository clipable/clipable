// Package modelsx contains developer-made types and objects (as opposed to sqlboiler-generated ones in models)
package modelsx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	log "github.com/sirupsen/logrus"
	"github.com/speps/go-hashids/v2"
	"github.com/volatiletech/null/v8"
)

// Unique error types
const (
	ErrorUniqueViolation  = "unique_violation"
	ErrorNotNullViolation = "not_null_violation"
)

var hashEncoder *hashids.HashID

func init() {
	SetHashEncoder("default salt")

	jsoniter.RegisterExtension(&NullExtension{})
	jsoniter.RegisterExtension(&CustomHashExtension{})

}

func SetHashEncoder(salt string) {
	var err error
	hashEncoder, err = hashids.NewWithData(&hashids.HashIDData{
		Alphabet:  hashids.DefaultAlphabet,
		MinLength: 4,
		Salt:      salt,
	})

	if err != nil {
		log.WithError(err).Fatal("Failed to create hash encoder")
	}
}

func HashEncode(id ...int64) (string, error) {
	return hashEncoder.EncodeInt64(id)
}

func HashDecodeSingle(encoded string) (int64, error) {
	nums, err := hashEncoder.DecodeInt64WithError(encoded)

	if err != nil {
		return 0, err
	}

	if len(nums) != 1 {
		return 0, errors.New("invalid hash")
	}

	return nums[0], nil
}

func HashDecode(encoded string) ([]int64, error) {
	return hashEncoder.DecodeInt64WithError(encoded)
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
	ret.SetTagName(tagKey)

	return ret
}

// TODO: This should be a uint64 but speps didn't feel like adding uint64 support https://github.com/speps/go-hashids/issues/21
type HashID int64

type CustomHashEncoder struct {
	Type reflect2.Type
}

func (e *CustomHashEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return e.Type.UnsafeIndirect(ptr).(HashID) == 0
}

func (e *CustomHashEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	id, err := HashEncode(int64(e.Type.UnsafeIndirect(ptr).(HashID)))

	if err != nil {
		log.WithError(err).Error("Failed to encode hash")
		return
	}

	stream.WriteString(id)
}

type CustomHashExtension struct {
	jsoniter.DummyExtension
}

func (ne *CustomHashExtension) CreateEncoder(t reflect2.Type) jsoniter.ValEncoder {
	if strings.HasPrefix(t.String(), "modelsx.HashID") {
		return &CustomHashEncoder{
			Type: t,
		}
	}

	return nil
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
