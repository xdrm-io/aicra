package dynfunc_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/aicra/internal/dynfunc"
)

func getReqType[Req, Res any](dynfunc.HandlerFunc[Req, Res]) reflect.Type {
	return reflect.TypeOf((*Req)(nil)).Elem()
}
func getResType[Req, Res any](dynfunc.HandlerFunc[Req, Res]) reflect.Type {
	return reflect.TypeOf((*Res)(nil)).Elem()
}

func testIn[Req any]() func(s *dynfunc.Signature) error {
	return func(s *dynfunc.Signature) error {
		var fn dynfunc.HandlerFunc[Req, struct{}] = func(context.Context, Req) (*struct{}, error) {
			return nil, nil
		}
		return s.ValidateRequest(getReqType(fn))
	}
}

func TestRequestValidation(t *testing.T) {
	t.Parallel()

	type CustomInt int
	type CustomFloat float64
	type CustomString string
	type CustomBytes []byte
	type CustomRunes []rune

	type User struct {
		ID       int
		Username string
		Email    string
	}

	tt := []struct {
		name   string
		config map[string]reflect.Type
		test   func(s *dynfunc.Signature) error
		err    error
	}{
		{
			name: "int",
			test: testIn[int](),
			err:  dynfunc.ErrNotAStruct,
		},
		{
			name: "struct pointer",
			test: testIn[*struct{}](),
			err:  dynfunc.ErrNotAStruct,
		},
		{
			name: "0 required 0 ok",
			test: testIn[struct{}](),
			err:  nil,
		},
		{
			name: "0 required 1 provided",
			test: testIn[struct{ ID int }](),
			err:  dynfunc.ErrUnexpectedFields,
		},

		{
			name: "1 required 0 provided",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{}](),
			err:  dynfunc.ErrMissingField,
		},
		{
			name: "1 required 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{ id int }](),
			err:  dynfunc.ErrUnexportedField,
		},
		{
			name: "1 required 1 invalid",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{ ID string }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 required 1 ok",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{ ID int }](),
			err:  nil,
		},
		{
			name: "1 required 1 struct ok",
			config: map[string]reflect.Type{
				"User": reflect.TypeOf(User{}),
			},
			test: testIn[struct{ User User }](),
			err:  nil,
		},
		{
			name: "1 required 1 type wrapper",
			config: map[string]reflect.Type{
				"Int": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{ Int CustomInt }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 required wrapper 1 primitive",
			config: map[string]reflect.Type{
				"Int": reflect.TypeOf(CustomInt(0)),
			},
			test: testIn[struct{ Int int }](),
			err:  dynfunc.ErrInvalidType,
		},

		{
			name: "1 optional 0 provided",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{}](),
			err:  dynfunc.ErrMissingField,
		},
		{
			name: "1 optional 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ id *int }](),
			err:  dynfunc.ErrUnexportedField,
		},
		{
			name: "1 optional 1 not pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ ID int }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 optional 1 invalid pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ ID *string }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 optional 1 ok",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ ID *int }](),
			err:  nil,
		},
		{
			name: "1 optional 1 struct ok",
			config: map[string]reflect.Type{
				"User": reflect.PointerTo(reflect.TypeOf(User{})),
			},
			test: testIn[struct{ User *User }](),
			err:  nil,
		},
		{
			name: "1 optional 1 type wrapper",
			config: map[string]reflect.Type{
				"Int": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ Int *CustomInt }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 optional wrapper 1 primitive",
			config: map[string]reflect.Type{
				"Int": reflect.PointerTo(reflect.TypeOf(CustomInt(0))),
			},
			test: testIn[struct{ Int *int }](),
			err:  dynfunc.ErrInvalidType,
		},

		{
			name: "N required 1 missing",
			config: map[string]reflect.Type{
				"Int":    reflect.TypeOf(int(0)),
				"Float":  reflect.TypeOf(float64(0)),
				"String": reflect.TypeOf(string("")),
				"Bytes":  reflect.TypeOf([]byte("")),
				"Runes":  reflect.TypeOf([]rune("")),
			},
			test: testIn[struct {
				Int   int
				Float float64
				// String string
				Bytes []byte
				Runes []rune
			}](),
			err: dynfunc.ErrMissingField,
		},
		{
			name: "N required N ok",
			config: map[string]reflect.Type{
				"Int":    reflect.TypeOf(int(0)),
				"Float":  reflect.TypeOf(float64(0)),
				"String": reflect.TypeOf(string("")),
				"Bytes":  reflect.TypeOf([]byte("")),
				"Runes":  reflect.TypeOf([]rune("")),
			},
			test: testIn[struct {
				Int    int
				Float  float64
				String string
				Bytes  []byte
				Runes  []rune
			}](),
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := &dynfunc.Signature{
				In: tc.config,
			}

			err := tc.test(s)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}
}

func testOut[Res any]() func(s *dynfunc.Signature) error {
	return func(s *dynfunc.Signature) error {
		var fn dynfunc.HandlerFunc[struct{}, Res] = func(context.Context, struct{}) (*Res, error) {
			return nil, nil
		}
		return s.ValidateResponse(getResType(fn))
	}
}

func TestResponseValidation(t *testing.T) {
	t.Parallel()

	type CustomInt int
	type CustomFloat float64
	type CustomString string
	type CustomBytes []byte
	type CustomRunes []rune

	type User struct {
		ID       int
		Username string
		Email    string
	}

	tt := []struct {
		name   string
		config map[string]reflect.Type
		test   func(s *dynfunc.Signature) error
		err    error
	}{
		{
			name: "int",
			test: testOut[int](),
			err:  dynfunc.ErrNotAStruct,
		},
		{
			name: "struct pointer",
			test: testOut[*struct{}](),
			err:  dynfunc.ErrNotAStruct,
		},
		{
			name: "0 required 0 ok",
			test: testOut[struct{}](),
			err:  nil,
		},
		{
			name: "0 required 1 provided",
			test: testOut[struct{ ID int }](),
			err:  dynfunc.ErrUnexpectedFields,
		},

		{
			name: "1 required 0 provided",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{}](),
			err:  dynfunc.ErrMissingField,
		},
		{
			name: "1 required 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{ id int }](),
			err:  dynfunc.ErrUnexportedField,
		},
		{
			name: "1 required 1 invalid",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{ ID string }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 required 1 ok",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{ ID int }](),
			err:  nil,
		},
		{
			name: "1 required 1 struct ok",
			config: map[string]reflect.Type{
				"User": reflect.TypeOf(User{}),
			},
			test: testOut[struct{ User User }](),
			err:  nil,
		},
		{
			name: "1 required 1 type wrapper",
			config: map[string]reflect.Type{
				"Int": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{ Int CustomInt }](),
			err:  nil,
		},
		{
			name: "1 required wrapper 1 primitive",
			config: map[string]reflect.Type{
				"Int": reflect.TypeOf(CustomInt(0)),
			},
			test: testOut[struct{ Int int }](),
			err:  nil,
		},

		{
			name: "1 optional 0 provided",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{}](),
			err:  dynfunc.ErrMissingField,
		},
		{
			name: "1 optional 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ id *int }](),
			err:  dynfunc.ErrUnexportedField,
		},
		{
			name: "1 optional 1 not pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ ID int }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 optional 1 invalid pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ ID *string }](),
			err:  dynfunc.ErrInvalidType,
		},
		{
			name: "1 optional 1 ok",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ ID *int }](),
			err:  nil,
		},
		{
			name: "1 optional 1 struct ok",
			config: map[string]reflect.Type{
				"User": reflect.PointerTo(reflect.TypeOf(User{})),
			},
			test: testOut[struct{ User *User }](),
			err:  nil,
		},
		{
			name: "1 optional 1 type wrapper",
			config: map[string]reflect.Type{
				"Int": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ Int *CustomInt }](),
			err:  nil,
		},
		{
			name: "1 optional wrapper 1 primitive",
			config: map[string]reflect.Type{
				"Int": reflect.PointerTo(reflect.TypeOf(CustomInt(0))),
			},
			test: testOut[struct{ Int *int }](),
			err:  nil,
		},

		{
			name: "N required 1 missing",
			config: map[string]reflect.Type{
				"Int":    reflect.TypeOf(int(0)),
				"Float":  reflect.TypeOf(float64(0)),
				"String": reflect.TypeOf(string("")),
				"Bytes":  reflect.TypeOf([]byte("")),
				"Runes":  reflect.TypeOf([]rune("")),
			},
			test: testOut[struct {
				Int   int
				Float float64
				// String string
				Bytes []byte
				Runes []rune
			}](),
			err: dynfunc.ErrMissingField,
		},
		{
			name: "N required N ok",
			config: map[string]reflect.Type{
				"Int":    reflect.TypeOf(int(0)),
				"Float":  reflect.TypeOf(float64(0)),
				"String": reflect.TypeOf(string("")),
				"Bytes":  reflect.TypeOf([]byte("")),
				"Runes":  reflect.TypeOf([]rune("")),
			},
			test: testOut[struct {
				Int    int
				Float  float64
				String string
				Bytes  []byte
				Runes  []rune
			}](),
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			s := &dynfunc.Signature{
				Out: tc.config,
			}

			err := tc.test(s)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}

}

func TestNewSignature(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		conf *config.Endpoint

		expect dynfunc.Signature
	}{
		{
			name: "ignore input without rename",
			conf: &config.Endpoint{
				Input: map[string]*config.Parameter{
					"Ignored": {GoType: reflect.TypeOf(int(0))},
					"Used":    {GoType: reflect.TypeOf(""), Rename: "Renamed"},
				},
			},
			expect: dynfunc.Signature{
				In: map[string]reflect.Type{
					"Renamed": reflect.TypeOf(""),
				},
			},
		},
		{
			name: "ignore output without rename",
			conf: &config.Endpoint{
				Output: map[string]*config.Parameter{
					"Ignored": {GoType: reflect.TypeOf(int(0))},
					"Used":    {GoType: reflect.TypeOf(""), Rename: "Renamed"},
				},
			},
			expect: dynfunc.Signature{
				Out: map[string]reflect.Type{
					"Renamed": reflect.TypeOf(""),
				},
			},
		},
		{
			name: "optional input pointers",
			conf: &config.Endpoint{
				Input: map[string]*config.Parameter{
					"Int":       {GoType: reflect.TypeOf(int(0)), Rename: "Int"},
					"OptInt":    {GoType: reflect.TypeOf(int(0)), Rename: "OptInt", Optional: true},
					"String":    {GoType: reflect.TypeOf(""), Rename: "String"},
					"OptString": {GoType: reflect.TypeOf(""), Rename: "OptString", Optional: true},
				},
			},
			expect: dynfunc.Signature{
				In: map[string]reflect.Type{
					"Int":       reflect.TypeOf(int(0)),
					"OptInt":    reflect.PointerTo(reflect.TypeOf(int(0))),
					"String":    reflect.TypeOf(""),
					"OptString": reflect.PointerTo(reflect.TypeOf("")),
				},
			},
		},
	}

	for _, tc := range tt {
		s := dynfunc.NewSignature(tc.conf)

		if len(s.In) != len(tc.expect.In) {
			t.Fatalf("invalid input count\nactual: %d\nexpect: %d", len(s.In), len(tc.expect.In))
		}
		for key, expect := range tc.expect.In {
			val, found := s.In[key]
			if !found {
				t.Fatalf("%q input missing", val)
			}
			if val.Kind() != expect.Kind() {
				t.Fatalf("invalid %1 input\nactual: %s\nexpect: %s", val, val.Kind(), expect.Kind())
			}
		}
		if len(s.Out) != len(tc.expect.Out) {
			t.Fatalf("invalid output count\nactual: %d\nexpect: %d", len(s.Out), len(tc.expect.Out))
		}
		for key, expect := range tc.expect.Out {
			val, found := s.Out[key]
			if !found {
				t.Fatalf("%q output missing", val)
			}
			if val.Kind() != expect.Kind() {
				t.Fatalf("invalid %1 output\nactual: %s\nexpect: %s", val, val.Kind(), expect.Kind())
			}
		}
	}
}
