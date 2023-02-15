package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"webserver/config"
	"webserver/services"
	"webserver/services/mock"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

func TestRoutes_SetProgress(t *testing.T) {
	tests := []struct {
		name     string
		group    *services.Group
		url      string
		payload  []byte
		expected int
		hasBody  bool
	}{
		{
			name:     "Success",
			expected: http.StatusOK,
			hasBody:  false,
			group: &services.Group{
				Transcoder: &mock.TranscoderProvider{
					ReportProgressHook: func(cid int64, frame int) {
						assert.Equal(t, frame, 686)
						assert.Equal(t, cid, int64(1))
					},
				},
			},
			url: "/progress/1",
			payload: []byte("frame=686\n" +
				"fps=74.23\n" +
				"stream_0_0_q=29.0\n" +
				"bitrate=N/A\n" +
				"total_size=N/A\n" +
				"out_time_us=19466732\n" +
				"out_time_ms=19466732\n" +
				"out_time=00:00:19.466732\n" +
				"dup_frames=0\n" +
				"drop_frames=682\n" +
				"speed=2.11x\n" +
				"progress=continue\n",
			),
		},
		{
			name:     "Handle invalid CID",
			expected: http.StatusBadRequest,
			hasBody:  true,
			url:      "/progress/invalid",
		},
		{
			name:     "Handle invalid line seperator",
			expected: http.StatusBadRequest,
			hasBody:  true,
			url:      "/progress/1",
			payload:  []byte("frame=686=123"),
		},
		{
			name:     "Handle invalid frame number",
			expected: http.StatusBadRequest,
			hasBody:  true,
			url:      "/progress/1",
			payload:  []byte("frame=invalid\nprogress=continue"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				Group: tt.group,
			}

			m := mux.NewRouter()
			m.HandleFunc("/progress/{cid}", r.SetProgress)
			req := httptest.NewRequest("POST", tt.url, bytes.NewReader(tt.payload))

			resp := httptest.NewRecorder()

			m.ServeHTTP(resp, req)
			if resp.Code != tt.expected {
				t.Errorf("Received unexpected error code during %s test. Wanted: %d Got: %d", tt.name, tt.expected, resp.Code)
			}

			if (resp.Body.Len() != 0) != tt.hasBody {
				t.Errorf("Received unexpected body during %s test. %s", tt.name, resp.Body.String())
			}
		})
	}
}

func TestRoutes_UploadObject(t *testing.T) {
	type fields struct {
		listeners      cmap.ConcurrentMap
		cfg            *config.Config
		Group          *services.Group
		store          sessions.Store
		Router         http.Handler
		InternalRouter http.Handler
	}
	type args struct {
		w   http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				listeners:      tt.fields.listeners,
				cfg:            tt.fields.cfg,
				Group:          tt.fields.Group,
				store:          tt.fields.store,
				Router:         tt.fields.Router,
				InternalRouter: tt.fields.InternalRouter,
			}
			r.UploadObject(tt.args.w, tt.args.req)
		})
	}
}

func TestRoutes_ReadObject(t *testing.T) {
	type fields struct {
		listeners      cmap.ConcurrentMap
		cfg            *config.Config
		Group          *services.Group
		store          sessions.Store
		Router         http.Handler
		InternalRouter http.Handler
	}
	type args struct {
		w   http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Routes{
				listeners:      tt.fields.listeners,
				cfg:            tt.fields.cfg,
				Group:          tt.fields.Group,
				store:          tt.fields.store,
				Router:         tt.fields.Router,
				InternalRouter: tt.fields.InternalRouter,
			}
			r.ReadObject(tt.args.w, tt.args.req)
		})
	}
}
