package runtime_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/runtime"
)

func TestRespond(t *testing.T) {
	tt := []struct {
		name string
		data map[string]any
		err  error

		wantStatus int
		wantBody   string
	}{
		{
			name:       "ok",
			data:       nil,
			err:        nil,
			wantStatus: 200,
			wantBody:   `{"status":"all right"}`,
		},
		{
			name:       "ok with data",
			data:       map[string]any{"foo": "bar"},
			err:        nil,
			wantStatus: 200,
			wantBody:   `{"foo":"bar","status":"all right"}`,
		},
		{
			name:       "ok with error",
			data:       nil,
			err:        api.Error(123, fmt.Errorf("-reason-")),
			wantStatus: 123,
			wantBody:   `{"status":"-reason-"}`,
		},
		{
			name:       "ok with data and error",
			data:       map[string]any{"foo": "bar"},
			err:        api.Error(123, fmt.Errorf("-reason-")),
			wantStatus: 123,
			wantBody:   `{"foo":"bar","status":"-reason-"}`,
		},
		{
			name:       "status conflict: error takes precedence",
			data:       map[string]any{"status": "foo"},
			err:        api.Error(123, fmt.Errorf("-reason-")),
			wantStatus: 123,
			wantBody:   `{"status":"-reason-"}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			w := httptest.NewRecorder()
			runtime.Respond(w, tc.data, tc.err)

			require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
			require.Equal(t, tc.wantStatus, w.Code, "http status code")

			require.Equal(t, tc.wantBody, w.Body.String(), "http body")
		})
	}
}
