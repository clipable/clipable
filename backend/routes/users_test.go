package routes

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"webserver/models"
	"webserver/modelsx"
	"webserver/services"
	"webserver/services/mock"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func TestRoutes_UpdateUser(t *testing.T) {
	tests := []struct {
		name     string
		group    *services.Group
		user     *models.User
		vars     *RouteVars
		payload  interface{}
		expected int
		hasBody  bool
		hasError bool
	}{
		{
			name:     "Success",
			expected: http.StatusOK,
			hasBody:  true,
			hasError: false,
			group: &services.Group{
				Users: &mock.UserProvider{
					UpdateHook: func(ctx context.Context, user *models.User, columns boil.Columns) error {
						return nil
					},
				},
			},
			user: &models.User{
				ID: 1,
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
		{
			name:     "Deny user updating another user",
			expected: http.StatusForbidden,
			hasBody:  false,
			hasError: false,
			group:    &services.Group{},
			user: &models.User{
				ID: 1,
			},
			vars: &RouteVars{
				UID: 2,
			},
		},
		{
			name:     "Deny when not authorized",
			expected: http.StatusUnauthorized,
			hasBody:  false,
			hasError: false,
			group:    &services.Group{},
			vars: &RouteVars{
				UID: 2,
			},
		},
		{
			name:     "Handle invalid body",
			expected: http.StatusBadRequest,
			hasBody:  true,
			hasError: false,
			payload: &modelsx.User{
				Username: null.StringFrom("1"),
			},
			group: &services.Group{},
			user: &models.User{
				ID: 1,
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
		{
			name:     "Handle update error",
			expected: http.StatusInternalServerError,
			hasBody:  false,
			hasError: true,
			group: &services.Group{
				Users: &mock.UserProvider{
					UpdateHook: func(ctx context.Context, user *models.User, columns boil.Columns) error {
						return sql.ErrConnDone
					},
				},
			},
			user: &models.User{
				ID: 1,
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				Group: tt.group,
			}

			body, _ := modelsx.UserSerialize.Marshal(tt.payload)

			req := httptest.NewRequest("GET", "/", bytes.NewReader(body))

			req = req.WithContext(context.WithValue(req.Context(), VarKey, tt.vars))

			code, body, err := r.UpdateUser(tt.user, req)
			if code != tt.expected {
				t.Errorf("Received unexpected error code during %s test. Wanted: %d Got: %d", tt.name, tt.expected, code)
			}

			if (body != nil) != tt.hasBody {
				t.Errorf("Received unexpected body during %s test.", tt.name)
			}

			if (err != nil) != tt.hasError {
				t.Errorf("Received unexpected error during %s test.", tt.name)
			}
		})
	}
}

func TestRoutes_SearchUsers(t *testing.T) {
	tests := []struct {
		name     string
		group    *services.Group
		user     *models.User
		vars     *RouteVars
		expected int
		hasBody  bool
		hasError bool
	}{
		{
			name:     "Success",
			expected: http.StatusOK,
			hasBody:  true,
			hasError: false,
			group: &services.Group{
				Users: &mock.UserProvider{
					SearchManyHook: func(ctx context.Context, query string) (models.UserSlice, error) {
						return models.UserSlice{&models.User{ID: 2}}, nil
					},
				},
			},
		},
		{
			name:     "Handle search many error",
			expected: http.StatusInternalServerError,
			hasBody:  false,
			hasError: true,
			group: &services.Group{
				Users: &mock.UserProvider{
					SearchManyHook: func(ctx context.Context, query string) (models.UserSlice, error) {
						return nil, sql.ErrConnDone
					},
				},
			},
		},
		{
			name:     "Handle no users found",
			expected: http.StatusNoContent,
			hasBody:  false,
			hasError: false,
			group: &services.Group{
				Users: &mock.UserProvider{
					SearchManyHook: func(ctx context.Context, query string) (models.UserSlice, error) {
						return models.UserSlice{}, nil
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				Group: tt.group,
			}

			req := httptest.NewRequest("GET", "/", &bytes.Buffer{})

			req = req.WithContext(context.WithValue(req.Context(), VarKey, tt.vars))

			code, body, err := r.SearchUsers(tt.user, req)
			if code != tt.expected {
				t.Errorf("Received unexpected error code during %s test. Wanted: %d Got: %d", tt.name, tt.expected, code)
			}

			if (body != nil) != tt.hasBody {
				t.Errorf("Received unexpected body during %s test.", tt.name)
			}

			if (err != nil) != tt.hasError {
				t.Errorf("Received unexpected error during %s test.", tt.name)
			}
		})
	}
}

func TestRoutes_GetUser(t *testing.T) {
	tests := []struct {
		name     string
		group    *services.Group
		user     *models.User
		vars     *RouteVars
		expected int
		hasBody  bool
		hasError bool
	}{
		{
			name:     "Success",
			expected: http.StatusOK,
			hasBody:  true,
			hasError: false,
			group: &services.Group{
				Users: &mock.UserProvider{
					FindHook: func(ctx context.Context, uid int64) (*models.User, error) {
						return &models.User{}, nil
					},
				},
			},
			user: &models.User{
				ID: 1,
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
		{
			name:     "Handle find error",
			expected: http.StatusInternalServerError,
			hasBody:  false,
			hasError: true,
			group: &services.Group{
				Users: &mock.UserProvider{
					FindHook: func(ctx context.Context, uid int64) (*models.User, error) {
						return nil, sql.ErrConnDone
					},
				},
			},
			user: &models.User{
				ID: 1,
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
		{
			name:     "Handle no user found",
			expected: http.StatusNotFound,
			hasBody:  false,
			hasError: false,
			group: &services.Group{
				Users: &mock.UserProvider{
					FindHook: func(ctx context.Context, uid int64) (*models.User, error) {
						return nil, sql.ErrNoRows
					},
				},
			},
			user: &models.User{
				ID: 1,
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				Group: tt.group,
			}

			req := httptest.NewRequest("GET", "/", &bytes.Buffer{})

			req = req.WithContext(context.WithValue(req.Context(), VarKey, tt.vars))

			code, body, err := r.GetUser(tt.user, req)
			if code != tt.expected {
				t.Errorf("Received unexpected error code during %s test. Wanted: %d Got: %d", tt.name, tt.expected, code)
			}

			if (body != nil) != tt.hasBody {
				t.Errorf("Received unexpected body during %s test.", tt.name)
			}

			if (err != nil) != tt.hasError {
				t.Errorf("Received unexpected error during %s test.", tt.name)
			}
		})
	}
}

func TestRoutes_GetCurrentUser(t *testing.T) {
	tests := []struct {
		name     string
		group    *services.Group
		user     *models.User
		vars     *RouteVars
		expected int
		hasBody  bool
		hasError bool
	}{
		{
			name:     "Success",
			expected: http.StatusOK,
			hasBody:  true,
			hasError: false,
			group:    &services.Group{},
			user: &models.User{
				ID: 1,
			},
		},
		{
			name:     "Deny when not authorized",
			expected: http.StatusUnauthorized,
			hasBody:  false,
			hasError: false,
			group:    &services.Group{},
			vars: &RouteVars{
				UID: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				Group: tt.group,
			}

			req := httptest.NewRequest("GET", "/", &bytes.Buffer{})

			req = req.WithContext(context.WithValue(req.Context(), VarKey, tt.vars))

			code, body, err := r.GetCurrentUser(tt.user, req)
			if code != tt.expected {
				t.Errorf("Received unexpected error code during %s test. Wanted: %d Got: %d", tt.name, tt.expected, code)
			}

			if (body != nil) != tt.hasBody {
				t.Errorf("Received unexpected body during %s test.", tt.name)
			}

			if (err != nil) != tt.hasError {
				t.Errorf("Received unexpected error during %s test.", tt.name)
			}
		})
	}
}

func TestRoutes_GetUsersClips(t *testing.T) {
	tests := []struct {
		name     string
		group    *services.Group
		user     *models.User
		vars     *RouteVars
		expected int
		hasBody  bool
		hasError bool
	}{
		{
			name:     "Success",
			expected: http.StatusOK,
			hasBody:  true,
			hasError: false,
			group: &services.Group{
				Clips: &mock.ClipsProvider{
					FindManyHook: func(ctx context.Context, user *models.User, mods ...qm.QueryMod) (models.ClipSlice, error) {
						return models.ClipSlice{&models.Clip{}}, nil
					},
				},
			},
			vars: &RouteVars{
				UID: 1,
			},
			user: &models.User{
				ID: 1,
			},
		},
		{
			name:     "Internal Server Error",
			expected: http.StatusInternalServerError,
			hasBody:  false,
			hasError: true,
			group: &services.Group{
				Clips: &mock.ClipsProvider{
					FindManyHook: func(ctx context.Context, user *models.User, mods ...qm.QueryMod) (models.ClipSlice, error) {
						return nil, sql.ErrNoRows
					},
				},
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
		{
			name:     "No Content",
			expected: http.StatusNoContent,
			hasBody:  false,
			hasError: false,
			group: &services.Group{
				Clips: &mock.ClipsProvider{
					FindManyHook: func(ctx context.Context, user *models.User, mods ...qm.QueryMod) (models.ClipSlice, error) {
						return models.ClipSlice{}, nil
					},
				},
			},
			vars: &RouteVars{
				UID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				Group: tt.group,
			}

			req := httptest.NewRequest("GET", "/", &bytes.Buffer{})

			req = req.WithContext(context.WithValue(req.Context(), VarKey, tt.vars))

			code, body, err := r.GetUsersClips(tt.user, req)
			if code != tt.expected {
				t.Errorf("Received unexpected error code during %s test. Wanted: %d Got: %d", tt.name, tt.expected, code)
			}

			if (body != nil) != tt.hasBody {
				t.Errorf("Received unexpected body during %s test.", tt.name)
			}

			if (err != nil) != tt.hasError {
				t.Errorf("Received unexpected error during %s test.", tt.name)
			}
		})
	}
}

func TestRoutes_GetUsers(t *testing.T) {
	tests := []struct {
		name     string
		group    *services.Group
		user     *models.User
		vars     *RouteVars
		expected int
		hasBody  bool
		hasError bool
	}{
		{
			name:     "Success",
			expected: http.StatusOK,
			hasBody:  true,
			hasError: false,
			group: &services.Group{
				Users: &mock.UserProvider{
					FindManyHook: func(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error) {
						return models.UserSlice{&models.User{ID: 2}}, nil
					},
				},
			},
		},
		{
			name:     "Handle FindMany error",
			expected: http.StatusInternalServerError,
			hasBody:  false,
			hasError: true,
			group: &services.Group{
				Users: &mock.UserProvider{
					FindManyHook: func(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error) {
						return nil, sql.ErrConnDone
					},
				},
			},
		},
		{
			name:     "Handle no users found",
			expected: http.StatusNoContent,
			hasBody:  false,
			hasError: false,
			group: &services.Group{
				Users: &mock.UserProvider{
					FindManyHook: func(ctx context.Context, mods ...qm.QueryMod) (models.UserSlice, error) {
						return models.UserSlice{}, nil
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				Group: tt.group,
			}

			req := httptest.NewRequest("GET", "/", &bytes.Buffer{})

			req = req.WithContext(context.WithValue(req.Context(), VarKey, tt.vars))

			code, body, err := r.GetUsers(tt.user, req)
			if code != tt.expected {
				t.Errorf("Received unexpected error code during %s test. Wanted: %d Got: %d", tt.name, tt.expected, code)
			}

			if (body != nil) != tt.hasBody {
				t.Errorf("Received unexpected body during %s test.", tt.name)
			}

			if (err != nil) != tt.hasError {
				t.Errorf("Received unexpected error during %s test.", tt.name)
			}
		})
	}
}
