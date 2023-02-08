package modelsx

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type Progress struct {
	Progress int `json:"progress"`
}

func (p *Progress) Marshal() (int, []byte, error) {
	data, err := jsoniter.Marshal(p)
	return http.StatusOK, data, err
}
