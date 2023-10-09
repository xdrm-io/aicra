package runtime_test

import (
	"io"
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

type Param struct {
	name      string
	extractor validator.ExtractFunc[any]
	value     any
}

func TestParseForm_JSON(t *testing.T) {
	tt := []struct {
		name string
		body string

		err         error
		checkParams func(t *testing.T, form runtime.Form)
	}{
		{
			name: "invalid json",
			body: "not-json",
			err:  runtime.ErrInvalidJSON,
		},
		{
			name: "empty json",
			body: "",
			err:  nil,
			checkParams: func(t *testing.T, form runtime.Form) {
				require.EqualValues(t, runtime.Form{}, form, "unexpected form values")
			},
		},
		{
			name: "nominal",
			body: `{"a":"string","b":1.2,"c":[2,3]}`,
			err:  nil,
			checkParams: func(t *testing.T, form runtime.Form) {
				a, err := runtime.ExtractForm[string](form, "a", validator.String{}.Validate(nil))
				require.NoError(t, err, "param 'a' is not available in the parsed form")
				require.EqualValues(t, "string", a, "unexpected value")

				b, err := runtime.ExtractForm[float64](form, "b", validator.Float{}.Validate(nil))
				require.NoError(t, err, "param 'b' is not available in the parsed form")
				require.EqualValues(t, float64(1.2), b, "unexpected value")

				_, err = runtime.ExtractForm[any](form, "c", validator.Any{}.Validate(nil))
				require.False(t, err == runtime.ErrMissingParam, "expected param 'c' to be available in the parsed form")
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
			if tc.checkParams != nil {
				tc.checkParams(t, form)
			}
		})
	}
}

type failReader struct{}

func (failReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestParseForm_URLEncoded(t *testing.T) {
	tt := []struct {
		name string
		body io.Reader

		err         error
		checkParams func(t *testing.T, form runtime.Form)
	}{
		{
			name: "invalid reader",
			body: failReader{},
			err:  runtime.ErrInvalidURLEncoded,
		},
		{
			name: "invalid format",
			body: strings.NewReader("a;b=1"),
			err:  runtime.ErrInvalidURLEncoded,
		},
		{
			name: "empty",
			body: strings.NewReader(""),
			err:  nil,
		},
		{
			name: "nominal",
			body: strings.NewReader(`a=string&b=1.2&c=2&c=3`),
			err:  nil,
			checkParams: func(t *testing.T, form runtime.Form) {
				a, err := runtime.ExtractForm[string](form, "a", validator.String{}.Validate(nil))
				require.NoError(t, err, "param %q is not available in the parsed form", "a")
				require.EqualValues(t, "string", a, "unexpected value")

				b, err := runtime.ExtractForm[float64](form, "b", validator.Float{}.Validate(nil))
				require.NoError(t, err, "param %q is not available in the parsed form", "b")
				require.EqualValues(t, float64(1.2), b, "unexpected value")

				c, err := runtime.ExtractForm[[]uint](form, "c", uintSliceValidator{}.Validate(nil))
				require.NoError(t, err, "param %q is not available in the parsed form", "c")
				require.EqualValues(t, []uint{2, 3}, c, "unexpected value")
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			req, err := http.NewRequest("POST", "", tc.body)
			require.NoError(t, err, "cannot create request")
			req.Header.Set("Content-Type", string(runtime.URLEncoded))

			form, err := runtime.ParseForm(req)
			if tc.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			if tc.checkParams != nil {
				tc.checkParams(t, form)
			}
		})
	}
}
