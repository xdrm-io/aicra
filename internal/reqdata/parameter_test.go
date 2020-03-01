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
