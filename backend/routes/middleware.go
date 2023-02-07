package routes

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webserver/models"

	"github.com/araddon/dateparse"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Auth injects the user of a request to the handler
func (r *Routes) Auth(handler func(u *models.User, r *http.Request) (int, []byte, error)) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		// TODO: This doeesn't work for some reason, so I'm just going to use the default timeout for now until we can figure out why it returns "feature not supported"
		//ctr := http.NewResponseController(resp)

		// fmt.Println(time.Now().Add(time.Duration(req.ContentLength/(500*KiB)) * time.Second).Add(1 * time.Second))
		// // Set a write timeout based on how long if would take to upload the body on a 500/kbps connection + 1 second for empty bodies
		// if err := ctr.SetReadDeadline(time.Now().Add(time.Duration(req.ContentLength/(500*KiB)) * time.Second).Add(1 * time.Second)); err != nil {
		// 	log.WithError(err).Errorln("Failed to set read deadline")
		// }
		// if err := ctr.SetWriteDeadline(time.Now().Add(time.Duration(req.ContentLength/(500*KiB)) * time.Second).Add(1 * time.Second)); err != nil {
		// 	log.WithError(err).Errorln("Failed to set write deadline")
		// }

		s, err := r.store.Get(req, SESSION_NAME)

		if err != nil {
			log.WithError(err).Errorln("Failed to get session")
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}

		// debug print s.Values
		// for k, v := range s.Values {
		// 	log.Debugf("s.Values[%s] = %v", k, v)
		// }

		var user *models.User

		raw, ok := s.Values[SESSION_KEY_ID]

		if ok {
			user, err = r.Users.Find(req.Context(), raw.(string))

			if err != nil {
				resp.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		code, body, err := handler(user, req)

		if req.Header.Get("X-CSRF-Token") == "" {
			resp.Header().Set("X-CSRF-Token", csrf.Token(req))
		}

		if body != nil && err == nil {
			resp.Header().Set("Content-Type", "application/json")
		}

		if e, ok := err.(net.Error); ok && e.Timeout() {
			code = http.StatusRequestTimeout
		}

		resp.WriteHeader(code)

		if code == http.StatusInternalServerError {
			body = []byte("Default Error Page")
		}

		if body != nil && err == nil {
			if _, err := resp.Write(body); err != nil {
				log.WithError(err).
					WithField("Path", req.URL.Path).
					Errorln("Failed to write body")
				return
			}
		}

		if err != nil {
			log.WithError(err).
				WithField("Path", req.URL.Path).
				WithField("Code", code).
				Warnln("Exception in route")
		}
	}
}

type RouteVars struct {
	UID string
	CID string
}

type key int

const VarKey = key(0)

func (r *Routes) ParseVars(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		rv := &RouteVars{}

		rv.UID, _ = vars["uid"]
		rv.CID, _ = vars["cid"]

		next.ServeHTTP(w, req.WithContext(
			context.WithValue(req.Context(), VarKey, rv),
		))
	})
}

func (r *Routes) SDCompliance(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "-1")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-Frame-Options", "DENY")
		w.Header().Add("frame-ancestors", "none")
		w.Header().Add("frame-src", "none")
		w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Add("Content-Security-Policy", "default-src 'none'; font-src 'none'; img-src 'self'; object-src 'none'; script-src 'self'")
		w.Header().Add("X-XSS-Protection", "0")
		next.ServeHTTP(w, req)
	})
}

func vars(r *http.Request) *RouteVars {
	return r.Context().Value(VarKey).(*RouteVars)
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
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	return lw.ResponseWriter.Write(b)
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
			"status": lw.statusCode,
			"method": r.Method,
			"took":   latency,
		}).Info(r.URL.Path)
	})
}
