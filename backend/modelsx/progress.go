package modelsx

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type Progress struct {
	Clips map[HashID]int `json:"clips"`
}

func (p *Progress) Marshal() (int, []byte, error) {
	data, err := jsoniter.Marshal(p)
	return http.StatusOK, data, err
}
