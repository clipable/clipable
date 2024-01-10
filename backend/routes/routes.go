// Package routes defines API endpoints
package routes

import (
	"crypto/tls"
	"database/sql"
	"net/http"
	"webserver/config"
	"webserver/services"
	"webserver/services/db"
	"webserver/services/object"
	"webserver/services/transcoder"

	"github.com/friendsofgo/errors"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/minio/minio-go/v7"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

// Routes contain pointers to resources needed in endpoint handlers
type Routes struct {
	listeners cmap.ConcurrentMap
	cfg       *config.Config
	*services.Group
	store sessions.Store

	Router         http.Handler
	InternalRouter http.Handler
}

// New configures the handler functions for each API endpoint
func New(cfg *config.Config, g *services.Group, store sessions.Store) (*Routes, error) {
	r := &Routes{
		listeners: cmap.New(),
		cfg:       cfg,
		Group:     g,
		store:     store,
	}

	logrus.SetLevel(logrus.InfoLevel)

	router := mux.NewRouter()
	internalRouter := mux.NewRouter()

	router.Use(DynamicTimeoutMiddleware)
	router.Use(LoggingMiddleware(logrus.InfoLevel))
	internalRouter.Use(LoggingMiddleware(logrus.DebugLevel))
	router.Use(r.ParseVars)
	if !cfg.Debug {
		router.Use(handlers.RecoveryHandler())
	}
	api := router.PathPrefix("/api").Subrouter()

	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{
			ServiceLabel: "clipable",
		}),
	})

	endpoint := func(path string, f http.HandlerFunc, method ...string) {
		api.Handle(path, std.Handler(path, mdlw, f)).Methods(method...)
	}

	internalEndpoint := func(path string, f http.HandlerFunc, method ...string) {
		internalRouter.Handle(path, f).Methods(method...)
	}

	// TODO: swap to https://github.com/uptrace/bunrouter?

	// INTERNAL ENDPOINTS
	internalEndpoint("/s3/{cid}/{file}", r.ReadObject, http.MethodGet)
	internalEndpoint("/s3/{cid}/{file}", r.UploadObject, http.MethodPost)
	internalEndpoint("/progress/{cid}", r.SetProgress, http.MethodPost)

	// AUTH ENDPOINTS
	endpoint("/auth/login", r.ResponseHandler(r.Login), http.MethodPost)
	endpoint("/auth/register", r.ResponseHandler(r.Register), http.MethodPost)
	endpoint("/auth/register", r.ResponseHandler(r.AllowRegistration), http.MethodOptions)
	endpoint("/auth/logout", r.ResponseHandler(r.Logout), http.MethodPost)

	// USER ENDPOINTS
	endpoint("/users/search", r.Handler(r.SearchUsers), http.MethodGet)
	endpoint("/users/me", r.Handler(r.GetCurrentUser), http.MethodGet)
	endpoint("/users", r.Handler(r.GetUsers), http.MethodGet)
	endpoint("/users/{uid:[a-zA-Z0-9-]{4,}}/clips", r.Handler(r.GetUsersClips), http.MethodGet)
	endpoint("/users/{uid:[a-zA-Z0-9-]{4,}}", r.Handler(r.GetUser), http.MethodGet)
	endpoint("/users/{uid:[a-zA-Z0-9-]{4,}}", r.Handler(r.UpdateUser), http.MethodPatch)

	// CLIP ENDPOINTS
	endpoint("/clips", r.Handler(r.UploadClip), http.MethodPost)
	endpoint("/clips", r.Handler(r.GetClips), http.MethodGet)
	endpoint("/clips/search", r.Handler(r.SearchClips), http.MethodGet)
	endpoint("/clips/progress", r.Handler(r.GetProgress), http.MethodGet)
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}", r.Handler(r.GetClip), http.MethodGet)
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}", r.Handler(r.UpdateClip), http.MethodPatch)
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}", r.Handler(r.DeleteClip), http.MethodDelete)

	// MPEG-DASH ENDPOINTS
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}/{filename}", r.StreamHandler(r.GetStreamFile), http.MethodGet)

	if cfg.CORS.Enabled {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		r.Router = cors.New(cors.Options{
			AllowedOrigins: []string{cfg.CORS.Origin, "https://reference.dashif.org", "https://shaka-player-demo.appspot.com", "https://csb-pygk8-mkhuda.vercel.app"},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler(router)
	} else {
		r.Router = router
	}

	r.InternalRouter = internalRouter

	return r, nil
}

// DefaultServiceGroup Comment for linter
func DefaultServiceGroup(cfg *config.Config, sdb *sql.DB, s3 *minio.Client) (*services.Group, error) {
	var err error
	group := &services.Group{
		Users:       db.NewUsers(sdb),
		ObjectStore: object.NewStore(s3, cfg),
	}

	group.Clips = db.NewClips(sdb, group.ObjectStore)
	group.Transcoder, err = transcoder.New(cfg, group)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create transcoder service")
	}

	return group, nil
}
