package api_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/ctx"
)

func TestContextExtract(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	ctx.Register(r, &api.Auth{})

	auth := api.Extract(r)
	require.NotNil(t, auth)
}
func TestContextNilExtract(t *testing.T) {
	fetched := api.Extract(nil)
	require.Nil(t, fetched)
}
func TestContextExtractNil(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	ctx.Register(r, nil)

	auth := api.Extract(r)
	require.Nil(t, auth)
}
func TestContextExtractInvalidType(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	ctx.Register(r, 123)

	auth := api.Extract(r)
	require.Nil(t, auth)
}
