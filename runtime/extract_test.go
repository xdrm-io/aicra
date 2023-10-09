package runtime_test

import (
	"net/http"
	"net/url"
	"strconv"
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

type uintSliceValidator struct{}

func (uintSliceValidator) Validate(params []string) validator.ExtractFunc[[]uint] {
	return func(value any) ([]uint, bool) {
		str, ok := value.([]string)
		if !ok {
			return []uint{}, false
		}
		cast := make([]uint, len(str))
		for i, v := range str {
			u, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return []uint{}, false
			}
			cast[i] = uint(u)
		}
		return cast, true
	}
}

func TestExtractQuery(t *testing.T) {
	tt := []struct {
		name      string
		req       *http.Request
		paramName string
		extractor validator.ExtractFunc[uint]

		err       error
		extracted any
	}{
		{
			name: "nil request",
			req:  nil,
			err:  runtime.ErrMissingParam,
		},
		{
			name: "invalid query",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "a;b=1",
				},
			},
			paramName: "a;b",
			err:       runtime.ErrMissingParam,
		},
		{
			name: "missing param",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "a=1",
				},
			},
			paramName: "b",
			err:       runtime.ErrMissingParam,
		},
		{
			name: "unexpected slice",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "a=1&a=2",
				},
			},
			paramName: "a",
			err:       runtime.ErrParseParameter,
		},
		{
			name: "invalid type",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "a=abc",
				},
			},
			paramName: "a",
			extractor: validator.Uint{}.Validate(nil),
			err:       runtime.ErrInvalidType,
		},
		{
			name: "ok",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "a=123",
				},
			},
			paramName: "a",
			extractor: validator.Uint{}.Validate(nil),
			extracted: uint(123),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			v, err := runtime.ExtractQuery(tc.req, tc.paramName, tc.extractor)
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

func TestExtractQuerySlice(t *testing.T) {
	tt := []struct {
		name      string
		req       *http.Request
		paramName string
		extractor validator.ExtractFunc[[]uint]

		err       error
		extracted any
	}{
		{
			name: "slice 1",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "a=1",
				},
			},
			paramName: "a",
			extractor: uintSliceValidator{}.Validate(nil),
			extracted: []uint{1},
		},
		{
			name: "slice 4",
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "a=1&a=2&a=3&a=4",
				},
			},
			paramName: "a",
			extractor: uintSliceValidator{}.Validate(nil),
			extracted: []uint{1, 2, 3, 4},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			v, err := runtime.ExtractQuery(tc.req, tc.paramName, tc.extractor)
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
