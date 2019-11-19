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

func TestStoreWithUri(t *testing.T) {
	urilist := []string{"abc", "def"}
	store := New(urilist, nil)

	if len(store.URI) != len(urilist) {
		t.Errorf("store 'Set' should contain %d elements (got %d)", len(urilist), len(store.URI))
		t.Fail()
	}
	if len(store.Set) != len(urilist) {
		t.Errorf("store 'Set' should contain %d elements (got %d)", len(urilist), len(store.Set))
		t.Fail()
	}

	for i, value := range urilist {

		t.Run(fmt.Sprintf("URL#%d='%s'", i, value), func(t *testing.T) {
			key := fmt.Sprintf("URL#%d", i)
			element, isset := store.Set[key]

			if !isset {
				t.Errorf("store should contain element with key '%s'", key)
				t.Failed()
			}

			if element.Value != value {
				t.Errorf("store[%s] should return '%s' (got '%s')", key, value, element.Value)
				t.Failed()
			}
		})

	}

}
