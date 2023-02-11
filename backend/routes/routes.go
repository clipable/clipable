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

	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/minio/minio-go/v7"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/cors"
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

	router := mux.NewRouter()
	internalRouter := mux.NewRouter()

	router.Use(LoggingMiddleware)
	internalRouter.Use(LoggingMiddleware)
	router.Use(r.ParseVars)
	if !cfg.Debug {
		router.Use(handlers.RecoveryHandler())
		router.Use(csrf.Protect([]byte(cfg.Cookie.Key), csrf.Path("/api")))
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
	internalEndpoint("/s3/{path}/{file}", r.ReadObject, http.MethodGet)
	internalEndpoint("/s3/{path}/{file}", r.UploadObject, http.MethodPost)

	// AUTH ENDPOINTS
	endpoint("/auth/login", r.Login, http.MethodPost)
	endpoint("/auth/register", r.Register, http.MethodPost)
	endpoint("/auth/logout", r.Logout, http.MethodPost)

	// USER ENDPOINTS
	endpoint("/users/search", r.Auth(r.SearchUsers), http.MethodGet)
	endpoint("/users/me", r.Auth(r.GetCurrentUser), http.MethodGet)
	endpoint("/users", r.Auth(r.GetUsers), http.MethodGet)
	endpoint("/users/{uid:[a-zA-Z0-9-]{4,}}", r.Auth(r.GetUser), http.MethodGet)
	endpoint("/users/{uid:[a-zA-Z0-9-]{4,}}", r.Auth(r.UpdateUser), http.MethodPatch)

	// CLIP ENDPOINTS
	endpoint("/clips", r.Auth(r.UploadClip), http.MethodPost)
	endpoint("/clips", r.Auth(r.GetClips), http.MethodGet)
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}", r.Auth(r.GetClip), http.MethodGet)
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}/progress", r.Auth(r.GetClipProgress), http.MethodGet)
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}", r.Auth(r.UpdateClip), http.MethodPatch)
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}", r.Auth(r.DeleteClip), http.MethodDelete)
	endpoint("/clips/search", r.Auth(r.SearchClips), http.MethodGet)

	// MPEG-DASH ENDPOINTS
	endpoint("/clips/{cid:[a-zA-Z0-9-]{4,}}/{filename}", r.GetStreamFile, http.MethodGet)

	if cfg.CORS.Enabled {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		r.Router = cors.New(cors.Options{
			AllowedOrigins: []string{cfg.CORS.Origin, "https://reference.dashif.org"},
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
	group := &services.Group{
		Users:       db.NewUsers(sdb),
		ObjectStore: object.NewStore(s3, cfg.S3.Bucket),
	}

	group.Clips = db.NewClips(sdb, group.ObjectStore)
	group.Transcoder = transcoder.New(group, 5)

	return group, nil
}
