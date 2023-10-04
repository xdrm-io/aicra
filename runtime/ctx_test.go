package runtime_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/ctx"
	"github.com/xdrm-io/aicra/runtime"
)

func TestGetAuthExtract(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	ctx.Register(r, &runtime.Context{
		Auth: &api.Auth{},
	})

	auth := runtime.GetAuth(r)
	require.NotNil(t, auth)
}
func TestGetAuthNilExtract(t *testing.T) {
	fetched := runtime.GetAuth(nil)
	require.Nil(t, fetched)
}
func TestGetAuthExtractNil(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	ctx.Register(r, nil)

	auth := runtime.GetAuth(r)
	require.Nil(t, auth)
}
func TestGetAuthExtractInvalidType(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	ctx.Register(r, 123)

	auth := runtime.GetAuth(r)
	require.Nil(t, auth)
}
