package dynfunc

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func getReqType[Req, Res any](HandlerFn[Req, Res]) reflect.Type {
	return reflect.TypeOf((*Req)(nil)).Elem()
}
func getResType[Req, Res any](HandlerFn[Req, Res]) reflect.Type {
	return reflect.TypeOf((*Res)(nil)).Elem()
}

func testIn[Req any]() func(s *Signature) error {
	return func(s *Signature) error {
		var fn HandlerFn[Req, struct{}] = func(context.Context, Req) (*struct{}, error) {
			return nil, nil
		}
		return s.ValidateRequest(getReqType(fn))
	}
}

func TestRequestValidation(t *testing.T) {

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
		test   func(s *Signature) error
		err    error
	}{
		{
			name: "int",
			test: testIn[int](),
			err:  ErrNotAStruct,
		},
		{
			name: "struct pointer",
			test: testIn[*struct{}](),
			err:  ErrNotAStruct,
		},
		{
			name: "0 required 0 ok",
			test: testIn[struct{}](),
			err:  nil,
		},
		{
			name: "0 required 1 provided",
			test: testIn[struct{ ID int }](),
			err:  ErrUnexpectedFields,
		},

		{
			name: "1 required 0 provided",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{}](),
			err:  ErrMissingField,
		},
		{
			name: "1 required 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{ id int }](),
			err:  ErrUnexportedField,
		},
		{
			name: "1 required 1 invalid",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testIn[struct{ ID string }](),
			err:  ErrInvalidType,
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
			err:  ErrInvalidType,
		},
		{
			name: "1 required wrapper 1 primitive",
			config: map[string]reflect.Type{
				"Int": reflect.TypeOf(CustomInt(0)),
			},
			test: testIn[struct{ Int int }](),
			err:  ErrInvalidType,
		},

		{
			name: "1 optional 0 provided",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{}](),
			err:  ErrMissingField,
		},
		{
			name: "1 optional 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ id *int }](),
			err:  ErrUnexportedField,
		},
		{
			name: "1 optional 1 not pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ ID int }](),
			err:  ErrInvalidType,
		},
		{
			name: "1 optional 1 invalid pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testIn[struct{ ID *string }](),
			err:  ErrInvalidType,
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
			err:  ErrInvalidType,
		},
		{
			name: "1 optional wrapper 1 primitive",
			config: map[string]reflect.Type{
				"Int": reflect.PointerTo(reflect.TypeOf(CustomInt(0))),
			},
			test: testIn[struct{ Int *int }](),
			err:  ErrInvalidType,
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
			err: ErrMissingField,
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
			s := &Signature{
				In: tc.config,
			}

			err := tc.test(s)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}
}

func testOut[Res any]() func(s *Signature) error {
	return func(s *Signature) error {
		var fn HandlerFn[struct{}, Res] = func(context.Context, struct{}) (*Res, error) {
			return nil, nil
		}
		return s.ValidateResponse(getResType(fn))
	}
}

func TestResponseValidation(t *testing.T) {
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
		test   func(s *Signature) error
		err    error
	}{
		{
			name: "int",
			test: testOut[int](),
			err:  ErrNotAStruct,
		},
		{
			name: "struct pointer",
			test: testOut[*struct{}](),
			err:  ErrNotAStruct,
		},
		{
			name: "0 required 0 ok",
			test: testOut[struct{}](),
			err:  nil,
		},
		{
			name: "0 required 1 provided",
			test: testOut[struct{ ID int }](),
			err:  ErrUnexpectedFields,
		},

		{
			name: "1 required 0 provided",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{}](),
			err:  ErrMissingField,
		},
		{
			name: "1 required 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{ id int }](),
			err:  ErrUnexportedField,
		},
		{
			name: "1 required 1 invalid",
			config: map[string]reflect.Type{
				"ID": reflect.TypeOf(int(0)),
			},
			test: testOut[struct{ ID string }](),
			err:  ErrInvalidType,
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
			err:  ErrMissingField,
		},
		{
			name: "1 optional 1 unexported",
			config: map[string]reflect.Type{
				"id": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ id *int }](),
			err:  ErrUnexportedField,
		},
		{
			name: "1 optional 1 not pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ ID int }](),
			err:  ErrInvalidType,
		},
		{
			name: "1 optional 1 invalid pointer",
			config: map[string]reflect.Type{
				"ID": reflect.PointerTo(reflect.TypeOf(int(0))),
			},
			test: testOut[struct{ ID *string }](),
			err:  ErrInvalidType,
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
			err: ErrMissingField,
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
			s := &Signature{
				Out: tc.config,
			}

			err := tc.test(s)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}

}
