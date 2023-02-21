package routes

import (
	"fmt"
	"io"
	"net/http"
	"webserver/models"

	"github.com/gotd/contrib/http_range"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func (r *Routes) GetStreamFile(w http.ResponseWriter, req *http.Request) {
	vars := vars(req)

	if !r.ObjectStore.HasObject(req.Context(), vars.CID, vars.Filename) {
		r.handleErr(w, http.StatusNotFound, nil, "Not Found")
		return
	}

	// Get the object from the minio server
	objReader, size, err := r.ObjectStore.GetObject(req.Context(), vars.CID, vars.Filename)

	if err != nil {
		r.handleErr(w, http.StatusInternalServerError, err, "Internal server error")
		return
	}

	defer objReader.Close()

	if vars.Filename == "dash.mpd" {
		// Get the clip to increment views by cid
		clip, err := r.Clips.Find(req.Context(), vars.CID)

		if err != nil {
			r.handleErr(w, http.StatusInternalServerError, err, "Internal server error")
			return
		}

		clip.Views++

		if err := r.Clips.Update(req.Context(), clip, boil.Whitelist(models.ClipColumns.Views)); err != nil {
			r.handleErr(w, http.StatusInternalServerError, err, "Internal server error")
			return
		}
	}

	ranges, err := http_range.ParseRange(req.Header.Get("Range"), size)

	if err != nil {
		r.handleErr(w, http.StatusBadRequest, err, "Bad Request")
		return
	}

	if len(ranges) > 1 {
		http.Error(w, "Requested Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
		r.handleErr(w, http.StatusRequestedRangeNotSatisfiable, nil, "Requested Range Not Satisfiable")
		return
	}

	if len(ranges) == 0 {
		// Set the content length
		w.Header().Set("Content-Length", fmt.Sprint(size))

		// Copy the object to the response writer
		_, err = io.Copy(w, objReader)
		if err != nil {
			r.handleErr(w, http.StatusInternalServerError, err, "Internal server error")
			return
		}

		return
	} else {
		// Accept ranges
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ranges[0].Start, ranges[0].Start+ranges[0].Length-1, size))
		w.Header().Set("Content-Length", fmt.Sprint(ranges[0].Length))

		// Seek to the start of the range
		_, err = objReader.Seek(ranges[0].Start, io.SeekStart)

		if err != nil {
			r.handleErr(w, http.StatusInternalServerError, err, "Internal server error")
			return
		}

		// Set the status code
		w.WriteHeader(http.StatusPartialContent)

		io.CopyN(w, objReader, ranges[0].Length)
		// TODO: Properly handle errors here and ignore if the error is due to the client disconnecting prematurely
	}
}
