package reqdata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmptyStore(t *testing.T) {
	store := New(nil, nil)

	if store.URI == nil {
		t.Errorf("store 'URI' list should be initialized")
		t.Fail()
	}
	if len(store.URI) != 0 {
		t.Errorf("store 'URI' list should be empty")
		t.Fail()
	}

	if store.Get == nil {
		t.Errorf("store 'Get' map should be initialized")
		t.Fail()
	}
	if store.Form == nil {
		t.Errorf("store 'Form' map should be initialized")
		t.Fail()
	}
	if store.Set == nil {
		t.Errorf("store 'Set' map should be initialized")
		t.Fail()
	}
}
