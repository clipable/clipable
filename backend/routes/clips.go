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
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (r *Routes) UploadClip(user *models.User, req *http.Request) (int, []byte, error) {
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

	clip.CreatorID = user.ID

	// Get the second part, which should be the video
	videoPart, err := mr.NextPart()

	if err == io.EOF {
		return http.StatusBadRequest, []byte("No video part"), nil
	}

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	// Check if the video part is a video
	if videoPart.FormName() != "video" {
		return http.StatusBadRequest, []byte("Second part must be video"), nil
	}

	model := clip.ToModel()

	// Create the clip
	tx, err := r.Clips.Create(req.Context(), model, boil.Whitelist(clip.GetUpdateWhitelist()...))

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	// Always attempt to rollback, even if it succeeds, if the tx is committed, this is a no-op
	defer tx.Rollback()

	len, err := tx.UploadVideo(req.Context(), io.LimitReader(videoPart, 2*GB))

	// LimitReader will return io.EOF once the limit is reached, so if we read exactly our limit
	// there was more data to read, and the video was too large
	if len == 2*GB {
		return http.StatusBadRequest, []byte("Video too large"), nil
	}

	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	if err := tx.Commit(); err != nil {
		return http.StatusInternalServerError, nil, err
	}

	return modelsx.ClipFromModel(model).Marshal()
}
