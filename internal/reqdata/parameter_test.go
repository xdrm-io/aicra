package reqdata

import (
	"testing"
)

func TestSimpleString(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: "some-string"}

	err := p.Parse()

	if err != nil {
		t.Errorf("unexpected error: <%s>", err)
		t.FailNow()
	}

	if !p.Parsed {
		t.Errorf("expected parameter to be parsed")
		t.FailNow()
	}

	cast, canCast := p.Value.(string)
	if !canCast {
		t.Errorf("expected parameter to be a string")
		t.FailNow()
	}

	if cast != "some-string" {
		t.Errorf("expected parameter to equal 'some-string', got '%s'", cast)
		t.FailNow()
	}
}
func TestStringSlice(t *testing.T) {
	p := Parameter{Parsed: false, File: false, Value: `["str1", "str2"]`}

	err := p.Parse()

	if err != nil {
		t.Errorf("unexpected error: <%s>", err)
		t.FailNow()
	}

	if !p.Parsed {
		t.Errorf("expected parameter to be parsed")
		t.FailNow()
	}

	slice, canCast := p.Value.([]interface{})
	if !canCast {
		t.Errorf("expected parameter to be a []interface{}")
		t.FailNow()
	}

	if len(slice) != 2 {
		t.Errorf("expected 2 values, got %d", len(slice))
		t.FailNow()
	}

	results := []string{"str1", "str2"}

	for i, res := range results {

		cast, canCast := slice[i].(string)
		if !canCast {
			t.Errorf("expected parameter %d to be a []string", i)
			continue
		}
		if cast != res {
			t.Errorf("expected first value to be '%s', got '%s'", res, cast)
			continue
		}

	}

}
