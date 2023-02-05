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
	Transcoder  Transcoder
	ObjectStore ObjectStore
	Users       Users
	Clips       Clips
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
	PutObject(id string, r io.Reader, size int64) (int64, error)
	GetObject(id string) (io.ReadSeekCloser, int64, error)

	DeleteObject(id string) error
	HasObject(id string) bool
}

// NewGroup Comment for linter
type Clips interface {
	Find(ctx context.Context, cid string) (*models.Clip, error)
	FindMany(ctx context.Context, mods ...qm.QueryMod) (models.ClipSlice, error)
	Exists(ctx context.Context, cid string) (bool, error)
	Delete(ctx context.Context, clip *models.Clip) error

	SearchMany(ctx context.Context, query string) (models.ClipSlice, error)

	Update(ctx context.Context, clip *models.Clip, columns boil.Columns) error
	Create(ctx context.Context, clip *models.Clip, creator *models.User, columns boil.Columns) (ClipTx, error)
}

type ClipTx interface {
	UploadVideo(ctx context.Context, r io.Reader) (int64, error)
	Commit() error
	Rollback() error
}

type Representations interface {
	Find(ctx context.Context, rid string) (*models.Representation, error)
	FindMany(ctx context.Context, mods ...qm.QueryMod) (models.RepresentationSlice, error)
	Exists(ctx context.Context, rid string) (bool, error)
	Delete(ctx context.Context, rep *models.Representation) error

	Update(ctx context.Context, rep *models.Representation, columns boil.Columns) error
	Create(ctx context.Context, rep *models.Representation, columns boil.Columns) error
}

type Transcoder interface {
	Queue(ctx context.Context, clip *models.Clip) error
}
