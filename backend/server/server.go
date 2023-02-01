// Package server defines the server object
package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
	"webserver/config"
	"webserver/routes"

	"github.com/gorilla/sessions"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/golang-migrate/migrate/v4"
	// Migrate Postgres driver import
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Migrate file driver import
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Postgres driver import
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	. "github.com/docker/go-units"
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
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	m, err := migrate.New(
		"file://migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&x-multi-statement=true", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name),
	)

	if err != nil {
		return nil, err
	}

	// Only uncomment this if you need to wipe the db
	// fmt.Println(m.Force(2))
	// fmt.Println(m.Down())
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	cookieStore := sessions.NewCookieStore([]byte(cfg.Cookie.Key), []byte(cfg.Cookie.Key))
	cookieStore.Options.SameSite = http.SameSiteLaxMode
	cookieStore.Options.Path = "/api"
	cookieStore.Options.Domain = cfg.Cookie.Domain
	cookieStore.MaxAge(int((30 * (24 * time.Hour)).Seconds())) // 30 Days

	s3, err := minio.New(cfg.S3.Address, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.Access, cfg.S3.Secret, ""),
		Secure: !cfg.Debug,
	})

	if err != nil {
		return nil, err
	}

	if s3.IsOffline() {
		return nil, errors.New("unable to connect to S3")
	}

	group, err := routes.DefaultServiceGroup(cfg, db, s3)

	if err != nil {
		return nil, err
	}

	r, err := routes.New(cfg, group, cookieStore)

	if err != nil {
		return nil, err
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

	srv := &http.Server{
		Addr:           s.cfg.ListenAddr,
		Handler:        s.routes.Router,
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 * MB,
	}
	return srv.ListenAndServe()
}
