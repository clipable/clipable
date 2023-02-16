package routes

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"webserver/models"
	"webserver/services"
	"webserver/services/mock"

	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func NewNopReadSeekCloser(body []byte) io.ReadSeekCloser {
	return &NopReadSeekCloser{body: bytes.NewReader(body)}
}

func NewErrorReadSeekCloser(seekError, readError error) io.ReadSeekCloser {
	return &NopReadSeekCloser{body: bytes.NewReader([]byte{}), seekError: seekError, readError: readError}
}

type NopReadSeekCloser struct {
	body      io.ReadSeeker
	seekError error
	readError error
}

func (n *NopReadSeekCloser) Read(p []byte) (int, error) {
	if n.readError != nil {
		return 0, n.readError
	}
	return n.body.Read(p)
}

func (n *NopReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	if n.seekError != nil {
		return 0, n.seekError
	}
	return n.body.Seek(offset, whence)
}

func (n *NopReadSeekCloser) Close() error {
	return nil
}

func TestRoutes_GetStreamFile(t *testing.T) {
	tests := []struct {
		name            string
		group           *services.Group
		vars            *RouteVars
		payload         []byte
		headers         map[string]string
		expected        int
		expectedHeaders map[string]string
		hasBody         bool
		bodyLength      int
	}{
		{
			name:       "Success - no range",
			expected:   http.StatusOK,
			hasBody:    true,
			bodyLength: 4,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return true
					},
					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewNopReadSeekCloser([]byte("test")), 4, nil
					},
				},
			},
		},
		{
			name:       "Success - range",
			expected:   http.StatusPartialContent,
			hasBody:    true,
			bodyLength: 3,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			headers: map[string]string{
				"Range": "bytes=0-2",
			},
			expectedHeaders: map[string]string{
				"Content-Range": "bytes 0-2/4",
			},

			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return true
					},
					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewNopReadSeekCloser([]byte("test")), 4, nil
					},
				},
			},
		},
		{
			name:       "Handle object doesn't exist",
			expected:   http.StatusNotFound,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return false
					},
				},
			},
		},
		{
			name:       "Handle error getting object",
			expected:   http.StatusInternalServerError,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return true
					},
					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return nil, 0, assert.AnError
					},
				},
			},
		},
		{
			name:       "Handle failure to find clip",
			expected:   http.StatusInternalServerError,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "dash.mpd",
			},
			group: &services.Group{
				Clips: &mock.ClipsProvider{
					FindHook: func(ctx context.Context, cid int64) (*models.Clip, error) {
						return nil, assert.AnError
					},
				},
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						return true
					},

					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewNopReadSeekCloser([]byte("test")), 4, nil
					},
				},
			},
		},
		{
			name:       "Handle failure to update clip view count",
			expected:   http.StatusInternalServerError,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "dash.mpd",
			},
			group: &services.Group{
				Clips: &mock.ClipsProvider{
					FindHook: func(ctx context.Context, cid int64) (*models.Clip, error) {
						return &models.Clip{}, nil
					},
					UpdateHook: func(ctx context.Context, clip *models.Clip, columns boil.Columns) error {
						return assert.AnError
					},
				},
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						return true
					},

					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewNopReadSeekCloser([]byte("test")), 4, nil
					},
				},
			},
		},
		{
			name:       "Handle failure to parse range header",
			expected:   http.StatusBadRequest,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			headers: map[string]string{
				"Range": "bytes=0-asd123",
			},
			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return true
					},
					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewNopReadSeekCloser([]byte("test")), 4, nil
					},
				},
			},
		},
		{
			name:       "Handle too many ranges",
			expected:   http.StatusRequestedRangeNotSatisfiable,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			headers: map[string]string{
				"Range": "bytes=0-1,2-3",
			},
			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return true
					},
					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewNopReadSeekCloser([]byte("test")), 4, nil
					},
				},
			},
		},
		{
			name:       "Handle failure to copy object",
			expected:   http.StatusInternalServerError,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return true
					},
					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewErrorReadSeekCloser(nil, assert.AnError), 4, nil
					},
				},
			},
		},
		{
			name:       "Handle failure to seek range",
			expected:   http.StatusInternalServerError,
			hasBody:    true,
			bodyLength: -1,
			vars: &RouteVars{
				CID:      1,
				Filename: "test.mp4",
			},
			headers: map[string]string{
				"Range": "bytes=0-1",
			},
			group: &services.Group{
				ObjectStore: &mock.ObjectStoreProvider{
					HasObjectHook: func(ctx context.Context, path string) bool {
						assert.Equal(t, "1/test.mp4", path)
						return true
					},
					GetObjectHook: func(ctx context.Context, path string) (io.ReadSeekCloser, int64, error) {
						return NewErrorReadSeekCloser(assert.AnError, nil), 4, nil
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

			req := httptest.NewRequest("GET", "/", bytes.NewReader(tt.payload))
			req = req.WithContext(context.WithValue(req.Context(), VarKey, tt.vars))

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			resp := httptest.NewRecorder()

			r.GetStreamFile(resp, req)

			assert.Equal(t, tt.expected, resp.Code)
			assert.Equal(t, tt.hasBody, resp.Body.Len() != 0)
			if tt.bodyLength != -1 {
				assert.Equal(t, tt.bodyLength, resp.Body.Len())
			}
			for k, v := range tt.expectedHeaders {
				assert.Equal(t, v, resp.Header().Get(k))
			}
		})
	}
}
