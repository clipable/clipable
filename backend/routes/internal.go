package routes

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gotd/contrib/http_range"
	log "github.com/sirupsen/logrus"
)

// Body looks like:
// frame=686
// fps=74.23
// stream_0_0_q=29.0
// bitrate=N/A
// total_size=N/A
// out_time_us=19466732
// out_time_ms=19466732
// out_time=00:00:19.466732
// dup_frames=0
// drop_frames=682
// speed=2.11x
// progress=continue
func (r *Routes) SetProgress(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	cid, err := strconv.ParseInt(vars["cid"], 10, 64)

	if err != nil {
		r.handleErr(w, http.StatusBadRequest, err, "Failed to parse cid")
		return
	}

	reader := bufio.NewScanner(req.Body)

	data := make(map[string]string)

	for reader.Scan() {
		line := reader.Text()

		parts := strings.Split(line, "=")

		if len(parts) != 2 {
			r.handleErr(w, http.StatusBadRequest, nil, "Bad Request")
			log.WithField("line", line).Error("Failed to parse line")
			return
		}

		data[parts[0]] = parts[1]

		if parts[0] == "progress" {
			// Get the current progress
			frame, err := strconv.Atoi(data["frame"])

			if err != nil {
				r.handleErr(w, http.StatusBadRequest, err, "Failed to parse frame")
				return
			}

			r.Transcoder.ReportProgress(cid, frame)
		}
	}
}

func (r *Routes) UploadObject(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	cid, err := strconv.ParseInt(vars["cid"], 10, 64)

	if err != nil {
		r.handleErr(w, http.StatusBadRequest, err, "Failed to parse cid")
		return
	}

	r.ObjectStore.PutObject(context.Background(), cid, vars["file"], req.Body)
}

func (r *Routes) ReadObject(w http.ResponseWriter, req *http.Request) {
	// Get the object ID from the URL
	vars := mux.Vars(req)
	cid, err := strconv.ParseInt(vars["cid"], 10, 64)

	if err != nil {
		r.handleErr(w, http.StatusBadRequest, err, "Failed to parse cid")
		return
	}

	if !r.ObjectStore.HasObject(context.Background(), cid, vars["file"]) {
		r.handleErr(w, http.StatusNotFound, nil, "Not Found")
		return
	}

	// Get the object from the minio server
	objReader, size, err := r.ObjectStore.GetObject(context.Background(), cid, vars["file"])

	if err != nil {
		r.handleErr(w, http.StatusInternalServerError, err, "Failed to get object")
		return
	}

	defer objReader.Close()

	ranges, err := http_range.ParseRange(req.Header.Get("Range"), size)

	if err != nil {
		r.handleErr(w, http.StatusBadRequest, err, "Bad Request")
		return
	}

	if len(ranges) > 1 {
		r.handleErr(w, http.StatusRequestedRangeNotSatisfiable, nil, "Requested Range Not Satisfiable")
		return
	}

	if len(ranges) == 1 {
		if ranges[0].Start > size || ranges[0].Start+ranges[0].Length > size {
			r.handleErr(w, http.StatusRequestedRangeNotSatisfiable, nil, "Requested Range Not Satisfiable")
			return
		}

		// Accept ranges
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ranges[0].Start, ranges[0].Start+ranges[0].Length-1, size))
		w.Header().Set("Content-Length", fmt.Sprint(ranges[0].Length))

		// Set the status code
		w.WriteHeader(http.StatusPartialContent)

		// Seek to the start of the range
		_, err = objReader.Seek(ranges[0].Start, io.SeekStart)

		if err != nil {
			r.handleErr(w, http.StatusInternalServerError, err, "Failed to seek to start of range")
			return
		}

		io.CopyN(w, objReader, ranges[0].Length)
	} else {
		w.Header().Set("Content-Length", fmt.Sprint(size))
		// Copy the object to the response writer
		_, err := io.Copy(w, objReader)
		if err != nil {
			r.handleErr(w, http.StatusInternalServerError, err, "Failed to copy object to response writer")
			return
		}
	}
}
