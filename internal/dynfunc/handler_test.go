package dynfunc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/xdrm-io/aicra/api"
	"github.com/xdrm-io/aicra/internal/config"
)

type fakeConfig config.Service

// builds a mock service with provided arguments as Input and matched as Output
func (s *fakeConfig) withArgs(dtypes ...reflect.Type) *fakeConfig {
	s.Input = make(map[string]*config.Parameter, len(dtypes))
	s.Output = make(map[string]*config.Parameter, len(dtypes))

	for i, dtype := range dtypes {
		name := fmt.Sprintf("P%d", i+1)
		s.Input[name] = &config.Parameter{
			Rename: name,
			GoType: dtype,
		}
		s.Output[name] = &config.Parameter{
			Rename: name,
			GoType: dtype,
		}
		if dtype.Kind() == reflect.Ptr {
			s.Output[name].GoType = dtype.Elem()
		}
	}
	return s
}

type fakeSign Signature

// builds a mock service with provided arguments as Input and matched as Output
func (s *fakeSign) withArgs(dtypes ...reflect.Type) *fakeSign {
	s.In = make(map[string]reflect.Type, len(dtypes))
	s.Out = make(map[string]reflect.Type, len(dtypes))

	for i, dtype := range dtypes {
		name := fmt.Sprintf("P%d", i+1)
		s.In[name] = dtype
		s.Out[name] = dtype
		if dtype.Kind() == reflect.Ptr {
			s.Out[name] = dtype.Elem()
		}
	}
	return s
}

func build[Req, Res any](fn HandlerFunc[Req, Res]) func(svc *config.Service) (Callable, error) {
	return func(svc *config.Service) (Callable, error) {
		return Build(svc, fn)
	}
}
func wrap[Req, Res any](fn HandlerFunc[Req, Res]) func(s *Signature) Callable {
	return func(s *Signature) Callable {
		return Wrap(s, fn)
	}
}

func TestInput(t *testing.T) {
	t.Parallel()

	type intstruct struct {
		P1 int
	}
	type intptrstruct struct {
		P1 *int
	}

	tt := []struct {
		name    string
		conf    *fakeConfig // warps config.Service
		hasCtx  bool
		builder func(svc *config.Service) (Callable, error)
		in      []interface{}
		out     []interface{}
		err     error
	}{
		{
			name: "none required none provided",
			conf: (&fakeConfig{}).withArgs(),
			builder: build(func(_ context.Context, r struct{}) (*struct{}, error) {
				return nil, nil
			}),
			hasCtx: false,
			in:     []interface{}{},
			out:    []interface{}{},
			err:    nil,
		},
		{
			name: "no input with error",
			conf: (&fakeConfig{}).withArgs(),
			builder: build(func(_ context.Context, r struct{}) (*struct{}, error) {
				return nil, api.ErrForbidden
			}),
			hasCtx: false,
			in:     []interface{}{},
			out:    []interface{}{},
			err:    api.ErrForbidden,
		},
		{
			name: "int proxy (0)",
			conf: (&fakeConfig{}).withArgs(reflect.TypeOf(int(0))),
			builder: build(func(_ context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, nil
			}),
			hasCtx: false,
			in:     []interface{}{int(0)},
			out:    []interface{}{int(0)},
			err:    nil,
		},
		{
			name: "int proxy with error",
			conf: (&fakeConfig{}).withArgs(reflect.TypeOf(int(0))),
			builder: build(func(ctx context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, api.ErrNotImplemented
			}),
			hasCtx: false,
			in:     []interface{}{int(0)},
			out:    []interface{}{int(0)},
			err:    api.ErrNotImplemented,
		},
		{
			name: "int proxy (11)",
			conf: (&fakeConfig{}).withArgs(reflect.TypeOf(int(0))),
			builder: build(func(ctx context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, nil
			}),
			hasCtx: false,
			in:     []interface{}{int(11)},
			out:    []interface{}{int(11)},
			err:    nil,
		},
		{
			name: "*int proxy (nil)",
			conf: (&fakeConfig{}).withArgs(reflect.TypeOf(new(int))),
			builder: build(func(ctx context.Context, in intptrstruct) (*intptrstruct, error) {
				return &intptrstruct{P1: in.P1}, nil
			}),
			hasCtx: false,
			in:     []interface{}{},
			out:    []interface{}{nil},
			err:    nil,
		},
		{
			name: "*int proxy (28)",
			conf: (&fakeConfig{}).withArgs(reflect.TypeOf(new(int))),
			builder: build(func(ctx context.Context, in intptrstruct) (*intstruct, error) {
				return &intstruct{P1: *in.P1}, nil
			}),
			hasCtx: false,
			in:     []interface{}{28},
			out:    []interface{}{28},
			err:    nil,
		},
		{
			name: "*int proxy (13)",
			conf: (&fakeConfig{}).withArgs(reflect.TypeOf(new(int))),
			builder: build(func(ctx context.Context, in intptrstruct) (*intstruct, error) {
				return &intstruct{P1: *in.P1}, nil
			}),
			hasCtx: false,
			in:     []interface{}{13},
			out:    []interface{}{13},
			err:    nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			callable, err := tc.builder((*config.Service)(tc.conf))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// build input
			input := make(map[string]interface{})
			for i, val := range tc.in {
				var key = fmt.Sprintf("P%d", i+1)
				input[key] = val
			}

			output, err := callable(context.Background(), input)
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
		builder  func(*Signature) Callable

		// do not expect a panic when empty
		panicMsg string

		// only checked when no panic happens
		err error
	}{
		{
			name:     "panic on unsettable input",
			expected: map[string]reflect.Type{"unexported": reflect.TypeOf(true)},
			input:    map[string]interface{}{"unexported": true},
			builder:  wrap[Unsettable, struct{}](func(context.Context, Unsettable) (*struct{}, error) { return nil, nil }),
			panicMsg: `cannot set field "unexported"`,
		},
		{
			name:     "panic on incompatible pointer",
			expected: map[string]reflect.Type{"NotABoolPointer": reflect.PtrTo(reflect.TypeOf(true))},
			input:    map[string]interface{}{"NotABoolPointer": true},
			builder:  wrap[IntPointer, struct{}](func(context.Context, IntPointer) (*struct{}, error) { return nil, nil }),
			panicMsg: `cannot convert bool into *int`,
		},
		{
			name:     "panic on incompatible type",
			expected: map[string]reflect.Type{"NotABool": reflect.TypeOf(true)},
			input:    map[string]interface{}{"NotABool": true},
			builder:  wrap[Int, struct{}](func(context.Context, Int) (*struct{}, error) { return nil, nil }),
			panicMsg: `cannot convert bool into int`,
		},
		{
			name: "skip missing input data",
			expected: map[string]reflect.Type{
				"String": reflect.TypeOf(""),
				"IntPtr": reflect.PtrTo(reflect.TypeOf(int(0))),
			},
			input: nil, // no input provided
			builder: wrap[Big, struct{}](func(ctx context.Context, in Big) (*struct{}, error) {
				if len(in.String) > 0 {
					t.Fatalf("expected string param to be skipped")
				}
				if in.IntPtr != nil {
					t.Fatalf("expected pointer param to be skipped")
				}
				return nil, nil
			}),
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
			builder: wrap[Big, struct{}](func(ctx context.Context, in Big) (*struct{}, error) {
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
				return nil, api.ErrNotImplemented
			}),
			err: api.ErrNotImplemented,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				s        = &Signature{In: tc.expected}
				callable = tc.builder(s)
			)

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
			_, err := callable(context.Background(), tc.input)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}

		})
	}
}

func TestBuild(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name    string
		service config.Service
		builder func(s *config.Service) (Callable, error)
		err     error
	}{
		{
			// input already tested in signature.ValidateInput
			name:    "unexpected input",
			service: config.Service{},
			builder: build(func(context.Context, struct{ Int int }) (*struct{}, error) { return nil, nil }),
			err:     ErrUnexpectedFields,
		},
		{
			// output already tested in signature.ValidateOutput
			name:    "unexpected output",
			service: config.Service{},
			builder: build(func(context.Context, struct{}) (*struct{ Int int }, error) { return nil, nil }),
			err:     ErrUnexpectedFields,
		},
		{
			name:    "valid empty service",
			service: config.Service{},
			builder: build(func(context.Context, struct{}) (*struct{}, error) { return nil, nil }),
			err:     nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.builder(&tc.service)
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}
}
