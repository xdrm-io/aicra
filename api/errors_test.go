package api_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/xdrm-io/aicra/api"
)

func TestError_ErrorStatusFormat(t *testing.T) {
	tt := []struct {
		name   string
		status int
		reason string
		err    api.Err
	}{
		{
			name:   "simple",
			status: 1,
			reason: "abc",
			err:    api.Err("1:abc"),
		},
		{
			name:   "with separators",
			status: 1,
			reason: "a:b:c",
			err:    api.Err("1:a:b:c"),
		},
		{
			name:   "negative status",
			status: -1,
			reason: "abc",
			err:    api.Err("-1:abc"),
		},
		{
			name:   "negative max",
			status: -700,
			reason: "abc",
			err:    api.Err("-700:abc"),
		},
		{
			name:   "positive max",
			status: 700,
			reason: "abc",
			err:    api.Err("700:abc"),
		},
		{
			name:   "invalid status",
			status: 500, // defaults to internal server error
			reason: "abc",
			err:    api.Err("xxx:abc"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.reason {
				t.Fatalf("expected Error() = '%s', got '%s'", tc.reason, tc.err.Error())
			}
			if tc.err.Status() != tc.status {
				t.Fatalf("expected Status() = %d, got %d", tc.status, tc.err.Status())
			}
		})
	}
}

type statusError int

func (s statusError) Error() string {
	return "err"
}
func (s statusError) Status() int {
	return int(s)
}

func TestError_GetStatusHelper(t *testing.T) {

	tt := []struct {
		name   string
		err    error
		status int
	}{
		{"ok", nil, http.StatusOK},
		{"it failed", api.ErrFailure, http.StatusInternalServerError},
		{"not found", api.ErrNotFound, http.StatusNotFound},
		{"already exists", api.ErrAlreadyExists, http.StatusInternalServerError},
		{"create error", api.ErrCreate, http.StatusInternalServerError},
		{"update error", api.ErrUpdate, http.StatusInternalServerError},
		{"delete error", api.ErrDelete, http.StatusInternalServerError},
		{"transactional error", api.ErrTransaction, http.StatusInternalServerError},
		{"unknown service", api.ErrUnknownService, http.StatusServiceUnavailable},
		{"uncallable service", api.ErrUncallableService, http.StatusServiceUnavailable},
		{"not implemented", api.ErrNotImplemented, http.StatusNotImplemented},
		{"unauthorized", api.ErrUnauthorized, http.StatusUnauthorized},
		{"forbidden", api.ErrForbidden, http.StatusForbidden},
		{"missing parameter", api.ErrMissingParam, http.StatusBadRequest},
		{"invalid parameter", api.ErrInvalidParam, http.StatusBadRequest},

		// implementing the `interface { Status() int }` is ok
		{"custom error", statusError(444), 444},

		// fallback to 500 if no `Status() int` method exists
		{"fallback to 500", errors.New("error"), http.StatusInternalServerError},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if api.GetErrorStatus(tc.err) != tc.status {
				t.Fatalf("invalid status %d, expected %d", api.GetErrorStatus(tc.err), tc.status)
			}
		})
	}

}
