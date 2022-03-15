package api_test

import (
	"context"
	"testing"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/ctx"
)

func TestContextExtract(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, ctx.Key, &api.Context{})

	// fetch from context
	fetched := api.Extract(c)
	if fetched == nil {
		t.Fatalf("fetched context must not be nil")
	}
}
func TestContextNilExtract(t *testing.T) {
	fetched := api.Extract(nil)
	if fetched == nil {
		t.Fatalf("fetched context must not be nil")
	}
}
func TestContextExtractNil(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, ctx.Key, nil)

	// fetch from context
	fetched := api.Extract(c)
	if fetched == nil {
		t.Fatalf("fetched context must not be nil")
	}
}
func TestContextExtractInvalidType(t *testing.T) {
	c := context.Background()
	c = context.WithValue(c, ctx.Key, 123)

	// fetch from context
	fetched := api.Extract(c)
	if fetched == nil {
		t.Fatalf("fetched context must not be nil")
	}
}
