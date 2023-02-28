package routes

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webserver/models"
	"webserver/modelsx"

	"github.com/araddon/dateparse"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	. "github.com/docker/go-units"
)

type LimitedReadCloser struct {
	io.Reader
	io.Closer
}

func NewLimitedReadCloser(rc io.ReadCloser, limit int64) io.ReadCloser {
	return &LimitedReadCloser{Reader: io.LimitReader(rc, limit), Closer: rc}
}

func StringToStream(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func (r *Routes) Handler(handler func(u *models.User, r *http.Request) (int, []byte, error)) http.HandlerFunc {
	return r.FullHandler(func(u *models.User, w http.ResponseWriter, r *http.Request) (int, io.ReadCloser, http.Header, error) {
		code, body, err := handler(u, r)
		return code, io.NopCloser(bytes.NewReader(body)), nil, err
	})
}

func (r *Routes) ResponseHandler(handler func(w http.ResponseWriter, r *http.Request) (int, []byte, error)) http.HandlerFunc {
	return r.FullHandler(func(u *models.User, w http.ResponseWriter, r *http.Request) (int, io.ReadCloser, http.Header, error) {
		code, _, err := handler(w, r)
		return code, nil, nil, err
	})
}

func (r *Routes) StreamHandler(handler func(u *models.User, r *http.Request) (int, io.ReadCloser, http.Header, error)) http.HandlerFunc {
	return r.FullHandler(func(u *models.User, w http.ResponseWriter, r *http.Request) (int, io.ReadCloser, http.Header, error) {
		code, body, headers, err := handler(u, r)
		return code, body, headers, err
	})
}

func (r *Routes) FullHandler(handler func(u *models.User, w http.ResponseWriter, r *http.Request) (int, io.ReadCloser, http.Header, error)) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		s, _ := r.store.Get(req, SESSION_NAME)

		var user *models.User
		var err error

		raw, ok := s.Values[SESSION_KEY_ID]

		if ok {
			user, err = r.Users.Find(req.Context(), raw.(int64))

			if err != nil {
				log.WithError(err).Warnln("Failed to find user, old cookie?")
				delete(s.Values, SESSION_KEY_ID)
				r.store.Save(req, resp, s)
			}
		}

		code, body, headers, err := handler(user, resp, req)

		// If the user did not send a CSRF token, send one back
		// TODO: Validate that this is the most idiomatic way of handling CSRF tokens
		if req.Header.Get("X-CSRF-Token") == "" {
			resp.Header().Set("X-CSRF-Token", csrf.Token(req))
		}

		// If the handler returned headers, add them to the response
		for k, v := range headers {
			for _, v := range v {
				resp.Header().Add(k, v)
			}
		}

		// If the error returned is a net.Error, check if it's a timeout
		if e, ok := err.(net.Error); ok && e.Timeout() {
			code = http.StatusRequestTimeout
		}

		resp.WriteHeader(code)

		if code == http.StatusInternalServerError {
			// If for some reason body isn't nil and we're returning an internal server error, close it
			if body != nil {
				body.Close()
			}

			body = io.NopCloser(bytes.NewReader([]byte("Default Error Page")))
		}

		if err != nil {
			log.WithError(err).
				WithField("Path", req.URL.Path).
				WithField("Code", code).
				Warnln("Exception in route")
		}

		if body != nil {
			defer body.Close()
			if _, err := io.Copy(resp, body); err != nil {
				log.WithError(err).
					WithField("Path", req.URL.Path).
					Errorln("Failed to write body")
				return
			}
		}
	}
}

type RouteVars struct {
	UID      int64
	CID      int64
	Filename string
}

type QueryVars struct {
	CID []int64
}

type key int

const VarKey = key(0)
const QueryKey = key(1)

func vars(r *http.Request) *RouteVars {
	return r.Context().Value(VarKey).(*RouteVars)
}

func query(r *http.Request) *QueryVars {
	return r.Context().Value(QueryKey).(*QueryVars)
}

func (r *Routes) ParseVars(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		var err error

		rv := &RouteVars{}

		if uid, ok := vars["uid"]; ok {
			rv.UID, err = modelsx.HashDecodeSingle(uid)

			if err != nil {
				log.WithError(err).Errorln("Failed to decode uid")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid UID"))
				return
			}
		}

		if cid, ok := vars["cid"]; ok {
			rv.CID, err = modelsx.HashDecodeSingle(cid)

			if err != nil {
				log.WithError(err).Errorln("Failed to decode cid")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid CID"))
				return
			}
		}

		if filename, ok := vars["filename"]; ok {
			rv.Filename = filename
		}

		qv := &QueryVars{}

		if cids, ok := req.URL.Query()["cid"]; ok {
			qv.CID = make([]int64, len(cids))

			for i, cid := range cids {
				qv.CID[i], err = modelsx.HashDecodeSingle(cid)

				if err != nil {
					log.WithError(err).WithField("CID", cid).Errorln("Failed to decode cid")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Invalid CID"))
					return
				}
			}
		}

		ctx := context.WithValue(req.Context(), VarKey, rv)
		ctx = context.WithValue(ctx, QueryKey, qv)

		next.ServeHTTP(w, req.WithContext(
			ctx,
		))
	})
}

func getPaginationMods(req *http.Request, paginationColumn, table, idColumn string) []qm.QueryMod {
	qms := make([]qm.QueryMod, 0)

	operation := "<"

	if order := req.URL.Query().Get("order"); order != "" {
		switch strings.ToLower(order) {
		case "asc":
			qms = append(qms, qm.OrderBy(paginationColumn))
			operation = ">"
		case "desc":
			qms = append(qms, qm.OrderBy(paginationColumn+" DESC"))
		}
	} else {
		qms = append(qms, qm.OrderBy(paginationColumn+" DESC"))
	}

	if index := req.URL.Query().Get("index"); index != "" {
		qms = append(qms, qm.Where(paginationColumn+operation+"(SELECT "+paginationColumn+" FROM \""+table+"\" WHERE "+idColumn+" = ?)", index))
	}

	if rawLimit := req.URL.Query().Get("limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)

		if err == nil {
			if limit > 200 {
				limit = 200
			}

			qms = append(qms, qm.Limit(limit))
		}
	} else {
		qms = append(qms, qm.Limit(200))
	}

	return qms
}

func getTimeRangeMods(req *http.Request, timeColumn string) []qm.QueryMod {
	qms := make([]qm.QueryMod, 0)

	if rawStart := req.URL.Query().Get("start"); rawStart != "" {
		start, err := dateparse.ParseAny(rawStart)

		if err == nil {
			qms = append(qms, qm.Where(timeColumn+" > ?", start))
		}
	}

	if rawEnd := req.URL.Query().Get("end"); rawEnd != "" {
		end, err := dateparse.ParseAny(rawEnd)

		if err == nil {
			qms = append(qms, qm.Where(timeColumn+" < ?", end))
		}
	}

	return qms
}

func realIP(req *http.Request) string {
	ra := req.RemoteAddr
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    uint64
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, 0}
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	len, err := lw.ResponseWriter.Write(b)

	lw.written += uint64(len)

	return len, err
}

func (lw *loggingResponseWriter) Unwrap() http.ResponseWriter {
	return lw.ResponseWriter
}

// Middleware implement mux middleware interface
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		entry := log.NewEntry(log.StandardLogger())
		start := time.Now()

		if reqID := r.Header.Get("X-Request-Id"); reqID != "" {
			entry = entry.WithField("requestId", reqID)
		}

		if remoteAddr := realIP(r); remoteAddr != "" {
			entry = entry.WithField("remoteAddr", remoteAddr)
		}

		lw := newLoggingResponseWriter(w)
		next.ServeHTTP(lw, r)

		latency := time.Since(start)

		entry.WithFields(log.Fields{
			"status":  lw.statusCode,
			"method":  r.Method,
			"written": humanize.Bytes(lw.written),
			"read":    humanize.Bytes(uint64(r.ContentLength)),
			"took":    latency.Round(time.Millisecond),
		}).Info(r.URL.Path)
	})
}

type timeoutResponseWriter struct {
	http.ResponseWriter
	rc *http.ResponseController
}

func (tw *timeoutResponseWriter) WriteHeader(code int) {
	length, err := strconv.Atoi(tw.Header().Get("Content-Length"))

	deadline := time.Now().Add(5 * time.Second)

	if err == nil {
		deadline = time.Now().Add(timeoutFromLength(int64(length)))
	}

	if err := tw.rc.SetWriteDeadline(deadline); err != nil {
		log.WithError(err).Errorln("Failed to set write deadline")
	}

	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutResponseWriter) Unwrap() http.ResponseWriter {
	return tw.ResponseWriter
}

func DynamicTimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctr := http.NewResponseController(w)

		deadline := time.Now().Add(timeoutFromLength(r.ContentLength))

		if err := ctr.SetReadDeadline(deadline); err != nil {
			log.WithError(err).Errorln("Failed to set read deadline")
		}

		// Write deadline also covers reading the request body, so we can set it to the same value until we know the response length
		if err := ctr.SetWriteDeadline(deadline); err != nil {
			log.WithError(err).Errorln("Failed to set write deadline")
		}

		next.ServeHTTP(&timeoutResponseWriter{w, ctr}, r)
	})
}

func timeoutFromLength(length int64) time.Duration {
	transferSpeed := 500 * KiB         // Set timeout based on 500kbps connection
	minimumDuration := 5 * time.Second // A minimum timeout duration in case there was no body
	return time.Duration(length/int64(transferSpeed))*time.Second + minimumDuration
}
