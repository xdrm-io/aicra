package runtime_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdrm-io/aicra/runtime"
	"github.com/xdrm-io/aicra/validator"
)

func TestParseForm_GET(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	require.NoError(t, err, "cannot create request")

	form, err := runtime.ParseForm(req)
	require.NoError(t, err, "cannot parse form")

	require.EqualValues(t, runtime.Form{}, form, "unexpected form values")
}

func TestParseForm_MissingContentType(t *testing.T) {
	req, err := http.NewRequest("POST", "", nil)
	require.NoError(t, err, "cannot create request")

	_, err = runtime.ParseForm(req)
	require.Error(t, err)
	require.ErrorIs(t, err, runtime.ErrUnhandledContentType)
}

func TestParseForm_InvalidContentType(t *testing.T) {
	req, err := http.NewRequest("POST", "", nil)
	require.NoError(t, err, "cannot create request")
	req.Header.Set("Content-Type", "application/xml")

	_, err = runtime.ParseForm(req)
	require.Error(t, err)
	require.ErrorIs(t, err, runtime.ErrUnhandledContentType)
}

func TestParseForm_JSON(t *testing.T) {
	tt := []struct {
		name string
		body string

		err    error
		params map[string]any
	}{
		{
			name: "invalid json",
			body: "not-json",
			err:  runtime.ErrInvalidJSON,
		},
		{
			name:   "empty json",
			body:   "",
			err:    nil,
			params: nil,
		},
		{
			name: "nominal",
			body: `{"a":"string","b":1.2, "c": ["d", 3.4]}`,
			err:  nil,
			params: map[string]any{
				"a": "string",
				"b": 1.2,
				"c": []any{"d", 3.4},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			req, err := http.NewRequest("POST", "", strings.NewReader(tc.body))
			require.NoError(t, err, "cannot create request")
			req.Header.Set("Content-Type", string(runtime.JSON))

			form, err := runtime.ParseForm(req)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			// check empty form
			if len(tc.params) == 0 {
				require.EqualValues(t, runtime.Form{}, form, "unexpected form values")
				return
			}

			for name, value := range tc.params {
				extracted, err := runtime.ExtractForm[any](form, name, validator.Any{}.Validate(nil))
				require.NoError(t, err, "param is not available in the parsed form")
				require.EqualValues(t, value, extracted, "unexpected value")
			}
		})
	}
}
