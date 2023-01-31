package services

import (
	"context"
	"io"
	"webserver/models"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Group Comment for linter
type Group struct {
	ObjectStore ObjectStore
	Users       Users
}

// Users Comment for linter
type Users interface {
	Find(ctx context.Context, uid string) (*models.User, error)
	FindMany(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error)
	Exists(ctx context.Context, uid string) (bool, error)

	SearchMany(ctx context.Context, query string) (models.UserSlice, error)

	Update(ctx context.Context, user *models.User, columns boil.Columns) error
	Create(ctx context.Context, user *models.User, columns boil.Columns) error
}

type ObjectStore interface {
	PutObject(id string, r io.Reader, size int64) error
	GetObject(id string) (io.ReadCloser, error)
	DeleteObject(id string) error
}
