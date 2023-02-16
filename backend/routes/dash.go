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

	fullPath := fmt.Sprintf("%d/%s", vars.CID, vars.Filename)

	if !r.ObjectStore.HasObject(req.Context(), fullPath) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Get the object from the minio server
	objReader, size, err := r.ObjectStore.GetObject(req.Context(), fullPath)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer objReader.Close()

	if vars.Filename == "dash.mpd" {
		// Get the clip to increment views by cid
		clip, err := r.Clips.Find(req.Context(), vars.CID)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		clip.Views++

		if err := r.Clips.Update(req.Context(), clip, boil.Whitelist(models.ClipColumns.Views)); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	ranges, err := http_range.ParseRange(req.Header.Get("Range"), size)

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if len(ranges) > 1 {
		http.Error(w, "Requested Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if len(ranges) == 0 {
		// Set the content length
		w.Header().Set("Content-Length", fmt.Sprint(size))

		// Copy the object to the response writer
		_, err = io.Copy(w, objReader)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	} else {
		// Accept ranges
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ranges[0].Start, ranges[0].Start+ranges[0].Length-1, size))

		// Seek to the start of the range
		_, err = objReader.Seek(ranges[0].Start, io.SeekStart)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the status code
		w.WriteHeader(http.StatusPartialContent)

		io.CopyN(w, objReader, ranges[0].Length)
		// TODO: Properly handle errors here and ignore if the error is due to the client disconnecting prematurely
	}
}
