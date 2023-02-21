package routes

import (
	"fmt"
	"io"
	"net/http"
	"webserver/models"

	"github.com/friendsofgo/errors"
	"github.com/gotd/contrib/http_range"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (r *Routes) GetStreamFile(u *models.User, req *http.Request) (int, io.ReadCloser, http.Header, error) {
	vars := vars(req)

	if !r.ObjectStore.HasObject(req.Context(), vars.CID, vars.Filename) {
		return http.StatusNotFound, nil, nil, nil
	}

	// Get the object from the minio server
	objReader, size, err := r.ObjectStore.GetObject(req.Context(), vars.CID, vars.Filename)

	if err != nil {
		return http.StatusInternalServerError, nil, nil, errors.Wrap(err, "failed to get object")
	}

	if vars.Filename == "dash.mpd" {
		// Get the clip to increment views by cid
		clip, err := r.Clips.Find(req.Context(), vars.CID)

		if err != nil {
			return http.StatusInternalServerError, nil, nil, errors.Wrap(err, "failed to find clip")
		}

		clip.Views++

		if err := r.Clips.Update(req.Context(), clip, boil.Whitelist(models.ClipColumns.Views)); err != nil {
			return http.StatusInternalServerError, nil, nil, errors.Wrap(err, "failed to update clip")
		}
	}

	ranges, err := http_range.ParseRange(req.Header.Get("Range"), size)

	if err != nil {
		return http.StatusBadRequest, StringToStream(errors.Wrap(err, "Invalid Range").Error()), nil, nil
	}

	if len(ranges) > 1 {
		return http.StatusRequestedRangeNotSatisfiable, StringToStream("Multiple ranges not supported"), nil, nil
	}

	headers := make(http.Header)

	if len(ranges) == 0 {
		// Set the content length
		headers.Set("Content-Length", fmt.Sprint(size))

		return http.StatusOK, objReader, headers, nil
	} else {
		// Accept ranges
		headers.Set("Accept-Ranges", "bytes")
		headers.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ranges[0].Start, ranges[0].Start+ranges[0].Length-1, size))
		headers.Set("Content-Length", fmt.Sprint(ranges[0].Length))

		// Seek to the start of the range
		_, err = objReader.Seek(ranges[0].Start, io.SeekStart)

		if err != nil {
			return http.StatusInternalServerError, nil, nil, errors.Wrap(err, "failed to seek to start of range")
		}

		return http.StatusPartialContent, NewLimitedReadCloser(objReader, ranges[0].Length), headers, nil
	}
}
