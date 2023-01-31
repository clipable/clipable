package routes

import (
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"webserver/models"
	"webserver/modelsx"

	. "github.com/docker/go-units"
)

func (r *Routes) UploadClip(user *models.User, req *http.Request) (int, []byte, error) {
	//vars := vars(req)

	// Get media type information from the content type header
	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))

	if err != nil {
		log.Fatal(err)
	}

	// Check if the media type is multipart
	if !strings.HasPrefix(mediaType, "multipart/") {
		return http.StatusBadRequest, []byte("Content-Type is not multipart"), nil
	}

	mr := multipart.NewReader(req.Body, params["boundary"])

	// Get the first part, which should be the json
	json, err := mr.NextPart()

	if err == io.EOF {
		return http.StatusBadRequest, []byte("No json part"), nil
	}

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	if json.FormName() != "json" {
		return http.StatusBadRequest, []byte("First part must be json"), nil
	}

	// Parse the json into a clip
	clip, err := modelsx.ParseClip(json)

	if err != nil {
		return http.StatusBadRequest, []byte(err.Error()), nil
	}

	videoPart, err := mr.NextPart()

	if err == io.EOF {
		return http.StatusBadRequest, []byte("No video part"), nil
	}

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	if videoPart.FormName() != "video" {
		return http.StatusBadRequest, []byte("Second part must be video"), nil
	}

	io.LimitReader(videoPart, 2*GB)

	return http.StatusOK, nil, nil
}
