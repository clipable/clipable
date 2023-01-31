package db

import (
	"context"
	"database/sql"
	"webserver/models"
	"webserver/services"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type users struct {
	db *sql.DB
}

// NewUsers Comment for linter
func NewUsers(db *sql.DB) services.Users {
	return &users{db}
}

func (u *users) Find(ctx context.Context, uid string) (*models.User, error) {
	return models.FindUser(ctx, u.db, uid)
}

func (u *users) FindMany(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error) {
	return models.Users(mods...).All(ctx, u.db)
}

func (u *users) Exists(ctx context.Context, uid string) (bool, error) {
	return models.UserExists(ctx, u.db, uid)
}

func (u *users) SearchMany(ctx context.Context, query string) (models.UserSlice, error) {
	return models.Users(
		// TODO: Figure out a good way to not replicate the `f_concat_ws` composite index declaration everywhere
		qm.Select("*"),
		qm.Where("f_concat_ws(' ', username, email, firstname, lastname) ILIKE ?", "%"+query+"%"),
		qm.OrderBy("f_concat_ws(' ', username, email, firstname, lastname) <-> ?", "%"+query+"%"),
		qm.Limit(10),
	).All(ctx, u.db)
}

func (u *users) Update(ctx context.Context, user *models.User, columns boil.Columns) error {
	_, err := user.Update(ctx, u.db, columns)

	return err
}

func (u *users) Create(ctx context.Context, user *models.User, columns boil.Columns) error {
	return user.Insert(ctx, u.db, columns)
}
