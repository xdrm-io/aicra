package aicra

import (
	"net/http/httptest"
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
		err  error
		data map[string]interface{}
		json string
	}{
		{
			name: "empty success response",
			err:  nil,
			data: map[string]interface{}{},
			json: `{"status":"all right"}`,
		},
		{
			name: "empty failure response",
			err:  api.ErrFailure,
			data: map[string]interface{}{},
			json: `{"status":"it failed"}`,
		},
		{
			name: "success with data before err",
			err:  nil,
			data: map[string]interface{}{"a": 12},
			json: `{"a":12,"status":"all right"}`,
		},
		{
			name: "success with data right before err",
			err:  nil,
			data: map[string]interface{}{"r": 12},
			json: `{"r":12,"status":"all right"}`,
		},
		{
			name: "success with data right after err",
			err:  nil,
			data: map[string]interface{}{"t": 12},
			json: `{"status":"all right","t":12}`,
		},
		{
			name: "success with data after err",
			err:  nil,
			data: map[string]interface{}{"z": 12},
			json: `{"status":"all right","z":12}`,
		},
		{
			name: "success with data around err",
			err:  nil,
			data: map[string]interface{}{"r": "before", "t": "after"},
			json: `{"r":"before","status":"all right","t":"after"}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			DefaultResponder(rec, tc.data, tc.err)

			if string(rec.Body.Bytes()) != tc.json {
				t.Fatalf("mismatching json:\nexpect: %v\nactual: %v", printEscaped(tc.json), printEscaped(string(rec.Body.Bytes())))
			}

		})
	}

}
