package runtime_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/internal/ctx"
	"github.com/xdrm-io/aicra/runtime"
	"github.com/xdrm-io/aicra/validator"
)

func TestExtractURI(t *testing.T) {
	tt := []struct {
		name      string
		ctx       *runtime.Context
		index     int
		extractor validator.ExtractFunc[uint]

		err       error
		extracted any
	}{
		{
			name: "no context",
			ctx:  nil,
			err:  runtime.ErrMissingURIParameter,
		},
		{
			name: "invalid context",
			ctx: &runtime.Context{
				Fragments: nil,
			},
			err: runtime.ErrMissingURIParameter,
		},
		{
			name: "invalid index",
			ctx: &runtime.Context{
				Fragments: []string{"base", "2"},
			},
			index: 2,
			err:   runtime.ErrMissingURIParameter,
		},
		{
			name: "invalid",
			ctx: &runtime.Context{
				Fragments: []string{"base", "abc"},
			},
			index:     1,
			extractor: validator.Uint{}.Validate(nil),
			err:       runtime.ErrInvalidType,
		},
		{
			name: "valid",
			ctx: &runtime.Context{
				Fragments: []string{"base", "123"},
			},
			index:     1,
			extractor: validator.Uint{}.Validate(nil),
			extracted: uint(123),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			req, err := http.NewRequest("GET", "", nil)
			require.NoError(t, err, "cannot create request")

			ctx.Register(req, tc.ctx)

			v, err := runtime.ExtractURI(req, tc.index, tc.extractor)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.extracted, v)
		})
	}
}
