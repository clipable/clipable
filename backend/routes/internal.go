package routes

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gotd/contrib/http_range"
)

func (r *Routes) UploadObject(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	r.ObjectStore.PutObject(context.Background(), vars["path"]+"/"+vars["file"], req.Body, -1)
}

func (r *Routes) ReadObject(w http.ResponseWriter, req *http.Request) {
	// Get the object ID from the URL
	vars := mux.Vars(req)

	// Get the object from the minio server
	objReader, size, err := r.ObjectStore.GetObject(context.Background(), vars["path"]+"/"+vars["file"])

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer objReader.Close()

	ranges, err := http_range.ParseRange(req.Header.Get("Range"), size)

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if len(ranges) > 1 {
		http.Error(w, "Requested Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if len(ranges) == 1 {
		// Accept ranges
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ranges[0].Start, ranges[0].Length, size))
		w.Header().Set("Content-Length", fmt.Sprint(ranges[0].Length))

		// Set the status code
		w.WriteHeader(http.StatusPartialContent)

		// Seek to the start of the range
		_, err = objReader.Seek(ranges[0].Start, io.SeekStart)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		io.CopyN(w, objReader, ranges[0].Length)
	} else {
		w.Header().Set("Content-Length", fmt.Sprint(size))
		// Copy the object to the response writer
		_, err := io.Copy(w, objReader)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
