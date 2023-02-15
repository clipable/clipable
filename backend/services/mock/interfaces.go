package mock

import (
	"context"
	"errors"
	"io"

	"webserver/models"
	"webserver/services"

	jsoniter "github.com/json-iterator/go"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BrokenCodec struct {
	jsoniter.API
}

var ErrBrokenCodec = errors.New("im a broken codec and can't do anything")

func (b *BrokenCodec) Marshal(v interface{}) ([]byte, error) {
	return nil, ErrBrokenCodec
}

func (b *BrokenCodec) Unmarshal(data []byte, v interface{}) error {
	return ErrBrokenCodec
}

type UserProvider struct {
	FindHook           func(ctx context.Context, uid int64) (*models.User, error)
	FindManyHook       func(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error)
	FindUsernameHook   func(ctx context.Context, username string) (*models.User, error)
	ExistsHook         func(ctx context.Context, uid int64) (bool, error)
	ExistsUsernameHook func(ctx context.Context, username string) (bool, error)
	SearchManyHook     func(ctx context.Context, query string) (models.UserSlice, error)
	UpdateHook         func(ctx context.Context, user *models.User, columns boil.Columns) error
	CreateHook         func(ctx context.Context, user *models.User, columns boil.Columns) error
}

func (m *UserProvider) Find(ctx context.Context, uid int64) (*models.User, error) {
	return m.FindHook(ctx, uid)
}
func (m *UserProvider) FindMany(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error) {
	return m.FindManyHook(ctx, mods...)
}
func (m *UserProvider) FindUsername(ctx context.Context, username string) (*models.User, error) {
	return m.FindUsernameHook(ctx, username)
}
func (m *UserProvider) Exists(ctx context.Context, uid int64) (bool, error) {
	return m.ExistsHook(ctx, uid)
}
func (m *UserProvider) ExistsUsername(ctx context.Context, username string) (bool, error) {
	return m.ExistsUsernameHook(ctx, username)
}
func (m *UserProvider) SearchMany(ctx context.Context, query string) (models.UserSlice, error) {
	return m.SearchManyHook(ctx, query)
}
func (m *UserProvider) Update(ctx context.Context, user *models.User, columns boil.Columns) error {
	return m.UpdateHook(ctx, user, columns)
}
func (m *UserProvider) Create(ctx context.Context, user *models.User, columns boil.Columns) error {
	return m.CreateHook(ctx, user, columns)
}

type ObjectStoreProvider struct {
	PutObjectHook    func(ctx context.Context, id string, r io.Reader, size int64) (int64, error)
	GetObjectHook    func(ctx context.Context, id string) (io.ReadSeekCloser, int64, error)
	DeleteObjectHook func(ctx context.Context, id string) error
	HasObjectHook    func(ctx context.Context, id string) bool
}

func (m *ObjectStoreProvider) PutObject(ctx context.Context, id string, r io.Reader, size int64) (int64, error) {
	return m.PutObjectHook(ctx, id, r, size)
}

func (m *ObjectStoreProvider) GetObject(ctx context.Context, id string) (io.ReadSeekCloser, int64, error) {
	return m.GetObjectHook(ctx, id)
}

func (m *ObjectStoreProvider) DeleteObject(ctx context.Context, id string) error {
	return m.DeleteObjectHook(ctx, id)
}

func (m *ObjectStoreProvider) HasObject(ctx context.Context, id string) bool {
	return m.HasObjectHook(ctx, id)
}

type ClipsProvider struct {
	FindHook       func(ctx context.Context, cid int64) (*models.Clip, error)
	FindManyHook   func(ctx context.Context, mods ...qm.QueryMod) (models.ClipSlice, error)
	ExistsHook     func(ctx context.Context, cid int64) (bool, error)
	DeleteHook     func(ctx context.Context, clip *models.Clip) error
	SearchManyHook func(ctx context.Context, query string) (models.ClipSlice, error)
	UpdateHook     func(ctx context.Context, clip *models.Clip, columns boil.Columns) error
	CreateHook     func(ctx context.Context, clip *models.Clip, creator *models.User, columns boil.Columns) (services.ClipTx, error)
}

func (m *ClipsProvider) Find(ctx context.Context, cid int64) (*models.Clip, error) {
	return m.FindHook(ctx, cid)
}

func (m *ClipsProvider) FindMany(ctx context.Context, mods ...qm.QueryMod) (models.ClipSlice, error) {
	return m.FindManyHook(ctx, mods...)
}

func (m *ClipsProvider) Exists(ctx context.Context, cid int64) (bool, error) {
	return m.ExistsHook(ctx, cid)
}

func (m *ClipsProvider) Delete(ctx context.Context, clip *models.Clip) error {
	return m.DeleteHook(ctx, clip)
}

func (m *ClipsProvider) SearchMany(ctx context.Context, query string) (models.ClipSlice, error) {
	return m.SearchManyHook(ctx, query)
}

func (m *ClipsProvider) Update(ctx context.Context, clip *models.Clip, columns boil.Columns) error {
	return m.UpdateHook(ctx, clip, columns)
}

func (m *ClipsProvider) Create(ctx context.Context, clip *models.Clip, creator *models.User, columns boil.Columns) (services.ClipTx, error) {
	return m.CreateHook(ctx, clip, creator, columns)
}

type ClipTxProvider struct {
	UploadVideoHook func(ctx context.Context, r io.Reader) (int64, error)
	CommitHook      func() error
	RollbackHook    func() error
}

func (m *ClipTxProvider) UploadVideo(ctx context.Context, r io.Reader) (int64, error) {
	return m.UploadVideoHook(ctx, r)
}

func (m *ClipTxProvider) Commit() error {
	return m.CommitHook()
}

func (m *ClipTxProvider) Rollback() error {
	return m.RollbackHook()
}

type TranscoderProvider struct {
	StartHook          func() error
	QueueHook          func(ctx context.Context, clip *models.Clip) error
	GetProgressHook    func(cid int64) (int, bool)
	ReportProgressHook func(cid int64, progress int)
}

func (m *TranscoderProvider) Start() error {
	return m.StartHook()
}

func (m *TranscoderProvider) Queue(ctx context.Context, clip *models.Clip) error {
	return m.QueueHook(ctx, clip)
}

func (m *TranscoderProvider) GetProgress(cid int64) (int, bool) {
	return m.GetProgressHook(cid)
}

func (m *TranscoderProvider) ReportProgress(cid int64, progress int) {
	m.ReportProgressHook(cid, progress)
}
