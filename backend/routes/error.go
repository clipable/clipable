package routes

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type errorResponse struct {
	Message string `json:"message"`
}

func (r *Routes) handleErr(resp http.ResponseWriter, statusCode int, err error, message string) {
	log.WithError(err).Error(message)
	resp.WriteHeader(statusCode)
	resp.Header().Set("Content-Type", "application/json")
	json.NewEncoder(resp).Encode(errorResponse{Message: message})
}
