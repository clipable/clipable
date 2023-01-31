package mock

import (
	"context"
	"errors"

	"webserver/models"

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
	FindHook       func(ctx context.Context, uid string) (*models.User, error)
	FindManyHook   func(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error)
	ExistsHook     func(ctx context.Context, uid string) (bool, error)
	SearchManyHook func(ctx context.Context, query string) (models.UserSlice, error)
	UpdateHook     func(ctx context.Context, user *models.User, columns boil.Columns) error
	CreateHook     func(ctx context.Context, user *models.User, columns boil.Columns) error
}

func (m *UserProvider) Find(ctx context.Context, uid string) (*models.User, error) {
	return m.FindHook(ctx, uid)
}
func (m *UserProvider) FindMany(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error) {
	return m.FindManyHook(ctx, mods...)
}
func (m *UserProvider) Exists(ctx context.Context, uid string) (bool, error) {
	return m.ExistsHook(ctx, uid)
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
