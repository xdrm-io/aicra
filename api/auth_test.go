package api

import (
	"testing"
)

func TestCombination(t *testing.T) {
	tcases := []struct {
		Name     string
		Required [][]string
		Active   []string
		Granted  bool
	}{
		{
			Name:     "no requirement none given",
			Required: [][]string{},
			Active:   []string{},
			Granted:  true,
		},
		{
			Name:     "empty requirements none given",
			Required: [][]string{{}},
			Active:   []string{},
			Granted:  true,
		},
		{
			Name:     "no requirement 1 given",
			Required: [][]string{},
			Active:   []string{"a"},
			Granted:  true,
		},
		{
			Name:     "no requirement some given",
			Required: [][]string{},
			Active:   []string{"a", "b"},
			Granted:  true,
		},

		{
			Name:     "1 required none given",
			Required: [][]string{{"a"}},
			Active:   []string{},
			Granted:  false,
		},
		{
			Name:     "1 required fulfilled",
			Required: [][]string{{"a"}},
			Active:   []string{"a"},
			Granted:  true,
		},
		{
			Name:     "1 required mismatch",
			Required: [][]string{{"a"}},
			Active:   []string{"b"},
			Granted:  false,
		},
		{
			Name:     "2 required none gien",
			Required: [][]string{{"a", "b"}},
			Active:   []string{},
			Granted:  false,
		},
		{
			Name:     "2 required other given",
			Required: [][]string{{"a", "b"}},
			Active:   []string{"c"},
			Granted:  false,
		},
		{
			Name:     "2 required one given",
			Required: [][]string{{"a", "b"}},
			Active:   []string{"a"},
			Granted:  false,
		},
		{
			Name:     "2 required fulfilled",
			Required: [][]string{{"a", "b"}},
			Active:   []string{"a", "b"},
			Granted:  true,
		},
		{
			Name:     "2 or 2 required first fulfilled",
			Required: [][]string{{"a", "b"}, {"c", "d"}},
			Active:   []string{"a", "b"},
			Granted:  true,
		},
		{
			Name:     "2 or 2 required second fulfilled",
			Required: [][]string{{"a", "b"}, {"c", "d"}},
			Active:   []string{"c", "d"},
			Granted:  true,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.Name, func(t *testing.T) {

			auth := Auth{
				Required: tcase.Required,
				Active:   tcase.Active,
			}

			// all right
			if tcase.Granted == auth.Granted() {
				return
			}

			if tcase.Granted && !auth.Granted() {
				t.Fatalf("expected granted authorization")
			}
			t.Fatalf("unexpected granted authorization")
		})
	}
}

func TestAuthNilActive(t *testing.T) {
	auth := Auth{
		Required: [][]string{{"a"}},
		Active:   nil,
	}
	if auth.Granted() {
		t.Fatalf("should not be granted as auth.Active is nil")
	}
}
