package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	tt := []struct {
		name         string
		method       string
		lenFragments int

		hash uint64
	}{
		{
			name:   "UNKNOWN no fragment",
			method: "UNKNOWN", lenFragments: 0,
			hash: 0,
		},
		{
			name:   "GET no fragment",
			method: "GET", lenFragments: 0,
			hash: 1,
		},
		{
			name:   "POST no fragment",
			method: "POST", lenFragments: 0,
			hash: 2,
		},
		{
			name:   "PUT no fragment",
			method: "PUT", lenFragments: 0,
			hash: 3,
		},
		{
			name:   "DELETE 1 fragment",
			method: "DELETE", lenFragments: 1,
			hash: 4 + 10,
		},

		{
			name:   "UNKNOWN 1 fragment",
			method: "UNKNOWN", lenFragments: 1,
			hash: 0 + 10,
		},
		{
			name:   "GET 1 fragment",
			method: "GET", lenFragments: 1,
			hash: 1 + 10,
		},
		{
			name:   "POST 1 fragment",
			method: "POST", lenFragments: 1,
			hash: 2 + 10,
		},
		{
			name:   "PUT 1 fragment",
			method: "PUT", lenFragments: 1,
			hash: 3 + 10,
		},
		{
			name:   "DELETE 1 fragment",
			method: "DELETE", lenFragments: 1,
			hash: 4 + 10,
		},

		{
			name:   "UNKNOWN 6 fragment",
			method: "UNKNOWN", lenFragments: 6,
			hash: 0 + 10*6,
		},
		{
			name:   "GET 6 fragment",
			method: "GET", lenFragments: 6,
			hash: 1 + 10*6,
		},
		{
			name:   "POST 6 fragment",
			method: "POST", lenFragments: 6,
			hash: 2 + 10*6,
		},
		{
			name:   "PUT 6 fragment",
			method: "PUT", lenFragments: 6,
			hash: 3 + 10*6,
		},
		{
			name:   "DELETE 6 fragment",
			method: "DELETE", lenFragments: 6,
			hash: 4 + 10*6,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			h := hash(tc.method, tc.lenFragments)
			require.Equal(t, tc.hash, h)
		})
	}
}
