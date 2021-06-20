package aicra

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xdrm-io/aicra/api"
)

func printEscaped(raw string) string {
	raw = strings.ReplaceAll(raw, "\n", "\\n")
	raw = strings.ReplaceAll(raw, "\r", "\\r")
	return raw
}

func TestResponseJSON(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		err  api.Err
		data map[string]interface{}
		json string
	}{
		{
			name: "empty success response",
			err:  api.ErrSuccess,
			data: map[string]interface{}{},
			json: `{"error":{"code":0,"reason":"all right"}}`,
		},
		{
			name: "empty failure response",
			err:  api.ErrFailure,
			data: map[string]interface{}{},
			json: `{"error":{"code":1,"reason":"it failed"}}`,
		},
		{
			name: "empty unknown error response",
			err:  api.ErrUnknown,
			data: map[string]interface{}{},
			json: `{"error":{"code":-1,"reason":"unknown error"}}`,
		},
		{
			name: "success with data before err",
			err:  api.ErrSuccess,
			data: map[string]interface{}{"a": 12},
			json: `{"a":12,"error":{"code":0,"reason":"all right"}}`,
		},
		{
			name: "success with data right before err",
			err:  api.ErrSuccess,
			data: map[string]interface{}{"e": 12},
			json: `{"e":12,"error":{"code":0,"reason":"all right"}}`,
		},
		{
			name: "success with data right after err",
			err:  api.ErrSuccess,
			data: map[string]interface{}{"f": 12},
			json: `{"error":{"code":0,"reason":"all right"},"f":12}`,
		},
		{
			name: "success with data after err",
			err:  api.ErrSuccess,
			data: map[string]interface{}{"z": 12},
			json: `{"error":{"code":0,"reason":"all right"},"z":12}`,
		},
		{
			name: "success with data around err",
			err:  api.ErrSuccess,
			data: map[string]interface{}{"d": "before", "f": "after"},
			json: `{"d":"before","error":{"code":0,"reason":"all right"},"f":"after"}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := newResponse().WithError(tc.err)
			for k, v := range tc.data {
				res.WithValue(k, v)
			}

			raw, err := json.Marshal(res)
			if err != nil {
				t.Fatalf("cannot marshal to json: %s", err)
			}

			if string(raw) != tc.json {
				t.Fatalf("mismatching json:\nexpect: %v\nactual: %v", printEscaped(tc.json), printEscaped(string(raw)))
			}

		})
	}

}
