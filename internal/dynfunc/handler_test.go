package dynfunc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/xdrm-io/aicra/api"
)

type testsignature Signature

// builds a mock service with provided arguments as Input and matched as Output
func (s *testsignature) withArgs(dtypes ...reflect.Type) *testsignature {
	if s.In == nil {
		s.In = make(map[string]reflect.Type)
	}
	if s.Out == nil {
		s.Out = make(map[string]reflect.Type)
	}

	for i, dtype := range dtypes {
		name := fmt.Sprintf("P%d", i+1)
		s.In[name] = dtype
		if dtype.Kind() == reflect.Ptr {
			s.Out[name] = dtype.Elem()
		} else {
			s.Out[name] = dtype
		}
	}
	return s
}

func TestInput(t *testing.T) {

	type intstruct struct {
		P1 int
	}
	type intptrstruct struct {
		P1 *int
	}

	tt := []struct {
		name   string
		spec   *testsignature
		hasCtx bool
		fn     interface{}
		in     []interface{}

		out []interface{}
		err error
	}{
		{
			name:   "none required none provided",
			spec:   (&testsignature{}).withArgs(),
			fn:     func(context.Context) (*struct{}, error) { return nil, nil },
			hasCtx: false,
			in:     []interface{}{},
			out:    []interface{}{},
			err:    nil,
		},
		{
			name:   "no input with error",
			spec:   (&testsignature{}).withArgs(),
			fn:     func(context.Context) (*struct{}, error) { return nil, api.ErrForbidden },
			hasCtx: false,
			in:     []interface{}{},
			out:    []interface{}{},
			err:    api.ErrForbidden,
		},
		{
			name: "int proxy (0)",
			spec: (&testsignature{}).withArgs(reflect.TypeOf(int(0))),
			fn: func(ctx context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, nil
			},
			hasCtx: false,
			in:     []interface{}{int(0)},
			out:    []interface{}{int(0)},
			err:    nil,
		},
		{
			name: "int proxy with error",
			spec: (&testsignature{}).withArgs(reflect.TypeOf(int(0))),
			fn: func(ctx context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, api.ErrNotImplemented
			},
			hasCtx: false,
			in:     []interface{}{int(0)},
			out:    []interface{}{int(0)},
			err:    api.ErrNotImplemented,
		},
		{
			name: "int proxy (11)",
			spec: (&testsignature{}).withArgs(reflect.TypeOf(int(0))),
			fn: func(ctx context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, nil
			},
			hasCtx: false,
			in:     []interface{}{int(11)},
			out:    []interface{}{int(11)},
			err:    nil,
		},
		{
			name: "*int proxy (nil)",
			spec: (&testsignature{}).withArgs(reflect.TypeOf(new(int))),
			fn: func(ctx context.Context, in intptrstruct) (*intptrstruct, error) {
				return &intptrstruct{P1: in.P1}, nil
			},
			hasCtx: false,
			in:     []interface{}{},
			out:    []interface{}{nil},
			err:    nil,
		},
		{
			name: "*int proxy (28)",
			spec: (&testsignature{}).withArgs(reflect.TypeOf(new(int))),
			fn: func(ctx context.Context, in intptrstruct) (*intstruct, error) {
				return &intstruct{P1: *in.P1}, nil
			},
			hasCtx: false,
			in:     []interface{}{28},
			out:    []interface{}{28},
			err:    nil,
		},
		{
			name: "*int proxy (13)",
			spec: (&testsignature{}).withArgs(reflect.TypeOf(new(int))),
			fn: func(ctx context.Context, in intptrstruct) (*intstruct, error) {
				return &intstruct{P1: *in.P1}, nil
			},
			hasCtx: false,
			in:     []interface{}{13},
			out:    []interface{}{13},
			err:    nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var handler = &Handler{
				signature: &Signature{In: tc.spec.In, Out: tc.spec.Out},
				fn:        tc.fn,
			}

			// build input
			input := make(map[string]interface{})
			for i, val := range tc.in {
				var key = fmt.Sprintf("P%d", i+1)
				input[key] = val
			}

			var output, err = handler.Handle(context.Background(), input)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}

			// check output
			for i, expect := range tc.out {
				var (
					key         = fmt.Sprintf("P%d", i+1)
					val, exists = output[key]
				)
				if !exists {
					t.Fatalf("missing output[%s]", key)
				}
				if expect != val {
					var (
						expectt   = reflect.ValueOf(expect)
						valt      = reflect.ValueOf(val)
						expectNil = !expectt.IsValid() || expectt.Kind() == reflect.Ptr && expectt.IsNil()
						valNil    = !valt.IsValid() || valt.Kind() == reflect.Ptr && valt.IsNil()
					)
					// ignore both nil
					if valNil && expectNil {
						continue
					}
					t.Fatalf("invalid output %q\nactual: (%T) %v\nexpect: (%T) %v", key, val, val, expect, expect)
				}
			}

		})
	}

}

func TestUnexpectedErrors(t *testing.T) {
	t.Parallel()

	type Unsettable struct {
		unexported bool
	}
	type IntPointer struct {
		NotABoolPointer *int
	}
	type Int struct {
		NotABool int
	}

	type Multi struct {
		String string
		IntPtr *int
	}
	type Big struct {
		String    string
		Int       int
		StringPtr *string
		IntPtr    *int
	}

	tt := []struct {
		name     string
		expected map[string]reflect.Type
		input    map[string]interface{}
		fn       interface{}

		// do not expect a panic when empty
		panicMsg string

		// only checked when no panic happens
		err error
	}{
		{
			name:     "panic on unsettable input",
			expected: map[string]reflect.Type{"unexported": reflect.TypeOf(true)},
			input:    map[string]interface{}{"unexported": true},
			fn:       func(context.Context, Unsettable) error { return nil },
			panicMsg: `cannot set field "unexported"`,
		},
		{
			name:     "panic on incompatible pointer",
			expected: map[string]reflect.Type{"NotABoolPointer": reflect.PtrTo(reflect.TypeOf(true))},
			input:    map[string]interface{}{"NotABoolPointer": true},
			fn:       func(context.Context, IntPointer) error { return nil },
			panicMsg: `cannot convert bool into *int`,
		},
		{
			name:     "panic on incompatible type",
			expected: map[string]reflect.Type{"NotABool": reflect.TypeOf(true)},
			input:    map[string]interface{}{"NotABool": true},
			fn:       func(context.Context, Int) error { return nil },
			panicMsg: `cannot convert bool into int`,
		},
		{
			name: "skip missing input data",
			expected: map[string]reflect.Type{
				"String": reflect.TypeOf(""),
				"IntPtr": reflect.PtrTo(reflect.TypeOf(int(0))),
			},
			input: nil, // no input provided
			fn: func(ctx context.Context, in Big) error {
				if len(in.String) > 0 {
					t.Fatalf("expected string param to be skipped")
				}
				if in.IntPtr != nil {
					t.Fatalf("expected pointer param to be skipped")
				}
				return nil
			},
		},
		{
			name: "valid input",
			expected: map[string]reflect.Type{
				"String":    reflect.TypeOf(""),
				"StringPtr": reflect.PtrTo(reflect.TypeOf("")),
				"Int":       reflect.TypeOf(int(0)),
				"IntPtr":    reflect.PtrTo(reflect.TypeOf(int(0))),
			},
			input: map[string]interface{}{
				"String":    "Some string!",
				"StringPtr": "Some other string!",
				"Int":       24680,
				"IntPtr":    -13579,
			},
			fn: func(ctx context.Context, in Big) error {
				if in.String != "Some string!" {
					t.Fatalf("invalid string\nactual: %q\nexpect: %q", in.String, "Some string!")
				}
				if in.Int != 24680 {
					t.Fatalf("invalid int\nactual: %d\nexpect: %d", in.Int, 24680)
				}
				if in.IntPtr == nil {
					t.Fatalf("unexpected nil pointer")
				}
				if *in.StringPtr != "Some other string!" {
					t.Fatalf("invalid int\nactual: %q\nexpect: %q", *in.StringPtr, "Some other string!")
				}
				if in.StringPtr == nil {
					t.Fatalf("unexpected nil pointer")
				}
				if *in.IntPtr != -13579 {
					t.Fatalf("invalid int\nactual: %d\nexpect: %d", *in.IntPtr, -13579)
				}
				return api.ErrNotImplemented
			},
			err: api.ErrNotImplemented,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			handler := Handler{
				signature: &Signature{In: tc.expected, Out: nil},
				fn:        tc.fn,
			}

			defer func() {
				expectPanic := len(tc.panicMsg) > 0

				r := recover()
				if r == nil && expectPanic {
					t.Fatalf("expected a panic: %q", tc.panicMsg)
				}
				if r != nil && !expectPanic {
					t.Fatalf("unexected panic: %v", r)
				}
				if expectPanic && fmt.Sprintf("%v", r) != tc.panicMsg {
					t.Fatalf("invalid panic message\nactual: %q\nexpect: %q", r, tc.panicMsg)
				}
			}()
			_, err := handler.Handle(context.Background(), tc.input)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}

		})
	}
}
