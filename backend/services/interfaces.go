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
	Find(ctx context.Context, uid int64) (*models.User, error)
	FindMany(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error)
	FindUsername(ctx context.Context, username string) (*models.User, error)

	Exists(ctx context.Context, uid int64) (bool, error)
	ExistsUsername(ctx context.Context, username string) (bool, error)

	SearchMany(ctx context.Context, query string) (models.UserSlice, error)

	Update(ctx context.Context, user *models.User, columns boil.Columns) error
	Create(ctx context.Context, user *models.User, columns boil.Columns) error
}

type ObjectStore interface {
	PutObject(ctx context.Context, cid int64, filename string, r io.Reader) (int64, error)
	GetObject(ctx context.Context, cid int64, filename string) (io.ReadSeekCloser, int64, string, error)

	DeleteObject(ctx context.Context, cid int64, filename string) error
	HasObject(ctx context.Context, cid int64, filename string) bool
	HasActiveUploads(ctx context.Context, cid int64) bool
}

// NewGroup Comment for linter
type Clips interface {
	Find(ctx context.Context, cid int64) (*models.Clip, error)
	FindMany(ctx context.Context, mods ...qm.QueryMod) (models.ClipSlice, error)
	Exists(ctx context.Context, cid int64) (bool, error)
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

type Transcoder interface {
	Start() error
	Queue(ctx context.Context, clip *models.Clip) error
	GetProgress(cid int64) (int, bool)
	ReportProgress(cid int64, frame int)
}
