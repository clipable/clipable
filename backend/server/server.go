// Package server defines the server object
package server

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"webserver/config"
	"webserver/modelsx"
	"webserver/routes"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"

	"github.com/golang-migrate/migrate/v4"
	// Migrate Postgres driver import
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Migrate file driver import
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Postgres driver import
	_ "github.com/jackc/pgx/v4/stdlib"

	. "github.com/docker/go-units"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Server objects contains private route and config details
type Server struct {
	routes *routes.Routes
	cfg    *config.Config
}

// New creates a server based on config.Config object
func New(cfg *config.Config) (*Server, error) {
	//boil.DebugMode = cfg.Debug
	db, err := sql.Open("pgx", fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=disable", cfg.DB.Name, cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password))

	if err != nil {
		return nil, errors.Wrap(err, "failed to open db connection")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping db")
	}

	log.SetFormatter(&log.TextFormatter{
		DisableQuote: true,
	})

	m, err := migrate.New(
		"file://migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&x-multi-statement=true", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name),
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create migrate object")
	}

	// Only uncomment this if you need to wipe the db
	// fmt.Println(m.Force(2))
	// fmt.Println(m.Down())
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, errors.Wrap(err, "failed to migrate db")
	}

	modelsx.SetHashEncoder(cfg.DB.IDHashKey)

	keyHash := pbkdf2.Key([]byte(cfg.Cookie.Key), nil, 600_000, 32, sha256.New)

	cookieStore := sessions.NewCookieStore(keyHash, keyHash)
	cookieStore.Options.SameSite = http.SameSiteStrictMode
	cookieStore.Options.Domain = cfg.Cookie.Domain
	cookieStore.MaxAge(int((30 * (24 * time.Hour)).Seconds())) // 30 Days

	s3, err := minio.New(cfg.S3.Address, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.Access, cfg.S3.Secret, ""),
		Secure: cfg.S3.Secure,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create s3 client")
	}

	if s3.IsOffline() {
		return nil, errors.New("unable to connect to S3")
	}

	group, err := routes.DefaultServiceGroup(cfg, db, s3)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create service group")
	}

	r, err := routes.New(cfg, group, cookieStore)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create routes")
	}

	return &Server{
		routes: r,
		cfg:    cfg,
	}, nil
}

// Start starts the hosting of routes on the address specified in cfg
func (s *Server) Start() error {
	log.Infoln("Listening on", s.cfg.ListenAddr, "and", s.cfg.MetricsListenAddr)

	go http.ListenAndServe(s.cfg.MetricsListenAddr, promhttp.Handler())
	go http.ListenAndServe("127.0.0.1:12786", s.routes.InternalRouter)
	go s.routes.Transcoder.Start()

	srv := &http.Server{
		Addr:              s.cfg.ListenAddr,
		Handler:           s.routes.Router,
		ReadTimeout:       6 * time.Hour,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 * MB,
	}
	return srv.ListenAndServe()
}
