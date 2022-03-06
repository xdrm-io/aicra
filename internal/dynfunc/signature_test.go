package dynfunc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/xdrm-io/aicra/internal/config"
)

func TestInputValidation(t *testing.T) {
	tt := []struct {
		name  string
		input map[string]reflect.Type
		fn    interface{}
		err   error
	}{
		{
			name:  "missing context",
			input: map[string]reflect.Type{},
			fn:    func() {},
			err:   ErrMissingHandlerContextArgument,
		},
		{
			name:  "invalid context",
			input: map[string]reflect.Type{},
			fn:    func(int) {},
			err:   ErrInvalidHandlerContextArgument,
		},
		{
			name:  "no input 0 given",
			input: map[string]reflect.Type{},
			fn:    func(context.Context) {},
			err:   nil,
		},
		{
			name:  "no input 1 given",
			input: map[string]reflect.Type{},
			fn:    func(context.Context, int) {},
			err:   ErrUnexpectedInput,
		},
		{
			name:  "no input 2 given",
			input: map[string]reflect.Type{},
			fn:    func(context.Context, int, string) {},
			err:   ErrUnexpectedInput,
		},
		{
			name: "1 input 0 given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func(context.Context) {},
			err: ErrMissingHandlerInputArgument,
		},
		{
			name: "1 input non-struct given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func(context.Context, int) {},
			err: ErrMissingParamArgument,
		},
		{
			name: "unexported input",
			input: map[string]reflect.Type{
				"test1": reflect.TypeOf(int(0)),
			},
			fn:  func(context.Context, struct{}) {},
			err: ErrUnexportedName,
		},
		{
			name: "1 input empty struct given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func(context.Context, struct{}) {},
			err: ErrMissingConfigArgument,
		},
		{
			name: "1 input invalid given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func(context.Context, struct{ Test1 string }) {},
			err: ErrWrongParamTypeFromConfig,
		},
		{
			name: "1 input valid given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func(context.Context, struct{ Test1 int }) {},
			err: nil,
		},
		{
			name: "1 input ptr empty struct given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			fn:  func(context.Context, struct{}) {},
			err: ErrMissingConfigArgument,
		},
		{
			name: "1 input ptr invalid given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			fn:  func(context.Context, struct{ Test1 string }) {},
			err: ErrWrongParamTypeFromConfig,
		},
		{
			name: "1 input ptr invalid ptr type given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			fn:  func(context.Context, struct{ Test1 *string }) {},
			err: ErrWrongParamTypeFromConfig,
		},
		{
			name: "1 input ptr valid given",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			fn:  func(context.Context, struct{ Test1 *int }) {},
			err: nil,
		},
		{
			name: "1 valid string",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(string("")),
			},
			fn:  func(context.Context, struct{ Test1 string }) {},
			err: nil,
		},
		{
			name: "1 valid uint",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(uint(0)),
			},
			fn:  func(context.Context, struct{ Test1 uint }) {},
			err: nil,
		},
		{
			name: "1 valid float64",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(float64(0)),
			},
			fn:  func(context.Context, struct{ Test1 float64 }) {},
			err: nil,
		},
		{
			name: "1 valid []byte",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf([]byte("")),
			},
			fn:  func(context.Context, struct{ Test1 []byte }) {},
			err: nil,
		},
		{
			name: "1 valid []rune",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf([]rune("")),
			},
			fn:  func(context.Context, struct{ Test1 []rune }) {},
			err: nil,
		},
		{
			name: "1 valid *string",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(string)),
			},
			fn:  func(context.Context, struct{ Test1 *string }) {},
			err: nil,
		},
		{
			name: "1 valid *uint",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(uint)),
			},
			fn:  func(context.Context, struct{ Test1 *uint }) {},
			err: nil,
		},
		{
			name: "1 valid *float64",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(float64)),
			},
			fn:  func(context.Context, struct{ Test1 *float64 }) {},
			err: nil,
		},
		{
			name: "1 valid *[]byte",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new([]byte)),
			},
			fn:  func(context.Context, struct{ Test1 *[]byte }) {},
			err: nil,
		},
		{
			name: "1 valid *[]rune",
			input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new([]rune)),
			},
			fn:  func(context.Context, struct{ Test1 *[]rune }) {},
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// mock spec
			s := Signature{
				Input:  tc.input,
				Output: nil,
			}

			err := s.ValidateInput(reflect.TypeOf(tc.fn))
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}
}

func TestOutputValidation(t *testing.T) {
	tt := []struct {
		name   string
		output map[string]reflect.Type
		fn     interface{}
		err    error
	}{
		{
			name:   "no output missing err",
			output: map[string]reflect.Type{},
			fn:     func() {},
			err:    ErrMissingHandlerErrorArgument,
		},
		{
			name:   "no output invalid err",
			output: map[string]reflect.Type{},
			fn:     func() bool { return true },
			err:    ErrInvalidHandlerErrorArgument,
		},
		{
			name:   "1 output none required",
			output: map[string]reflect.Type{},
			fn:     func(context.Context) (*struct{}, error) { return nil, nil },
			err:    nil,
		},
		{
			name: "no output 1 required",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func() error { return nil },
			err: ErrMissingHandlerOutputArgument,
		},
		{
			name: "invalid int output",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func() (int, error) { return 0, nil },
			err: ErrWrongOutputArgumentType,
		},
		{
			name: "invalid int ptr output",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func() (*int, error) { return nil, nil },
			err: ErrWrongOutputArgumentType,
		},
		{
			name: "invalid struct output",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func() (struct{ Test1 int }, error) { return struct{ Test1 int }{Test1: 1}, nil },
			err: ErrWrongOutputArgumentType,
		},
		{
			name: "unexported param",
			output: map[string]reflect.Type{
				"test1": reflect.TypeOf(int(0)),
			},
			fn:  func() (*struct{}, error) { return nil, nil },
			err: ErrUnexportedName,
		},
		{
			name: "missing output param",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func() (*struct{}, error) { return nil, nil },
			err: ErrMissingConfigArgument,
		},
		{
			name: "invalid output param",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func() (*struct{ Test1 string }, error) { return nil, nil },
			err: ErrWrongParamTypeFromConfig,
		},
		{
			name: "valid param",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			fn:  func() (*struct{ Test1 int }, error) { return nil, nil },
			err: nil,
		},
		{
			name: "2 valid params",
			output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
				"Test2": reflect.TypeOf(string("")),
			},
			fn: func() (*struct {
				Test1 int
				Test2 string
			}, error) {
				return nil, nil
			},
			err: nil,
		},
		{
			name: "nil type ignore typecheck",
			output: map[string]reflect.Type{
				"Test1": nil,
			},
			fn:  func() (*struct{ Test1 int }, error) { return nil, nil },
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// mock spec
			s := Signature{
				Input:  nil,
				Output: tc.output,
			}
			err := s.ValidateOutput(reflect.TypeOf(tc.fn))
			if !errors.Is(err, tc.err) {
				t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
			}
		})
	}
}

func TestServiceValidation(t *testing.T) {

	tt := []struct {
		name string
		in   []*config.Parameter
		out  []*config.Parameter
		fn   interface{}
		err  error
	}{
		{
			name: "missing context",
			fn:   func() {},
			err:  ErrMissingHandlerContextArgument,
		},
		{
			name: "invalid context",
			fn:   func(int) {},
			err:  ErrInvalidHandlerContextArgument,
		},
		{
			name: "missing error",
			fn:   func(context.Context) {},
			err:  ErrMissingHandlerErrorArgument,
		},
		{
			name: "invalid error",
			fn:   func(context.Context) int { return 1 },
			err:  ErrInvalidHandlerErrorArgument,
		},
		{
			name: "no in no out",
			fn:   func(context.Context) error { return nil },
			err:  nil,
		},
		{
			name: "unamed in",
			in: []*config.Parameter{
				{
					Rename: "", // should be ignored
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context) error { return nil },
			err: nil,
		},
		{
			name: "missing in",
			in: []*config.Parameter{
				{
					Rename: "Test1",
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context) error { return nil },
			err: ErrMissingHandlerInputArgument,
		},
		{
			name: "valid in",
			in: []*config.Parameter{
				{
					Rename: "Test1",
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context, struct{ Test1 int }) error { return nil },
			err: nil,
		},
		{
			name: "optional in not ptr",
			in: []*config.Parameter{
				{
					Rename:   "Test1",
					GoType:   reflect.TypeOf(int(0)),
					Optional: true,
				},
			},
			fn:  func(context.Context, struct{ Test1 int }) error { return nil },
			err: ErrWrongParamTypeFromConfig,
		},
		{
			name: "valid optional in",
			in: []*config.Parameter{
				{
					Rename:   "Test1",
					GoType:   reflect.TypeOf(int(0)),
					Optional: true,
				},
			},
			fn:  func(context.Context, struct{ Test1 *int }) error { return nil },
			err: nil,
		},

		{
			name: "unamed out",
			out: []*config.Parameter{
				{
					Rename: "", // should be ignored
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context) error { return nil },
			err: nil,
		},
		{
			name: "missing out struct",
			out: []*config.Parameter{
				{
					Rename: "Test1",
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context) error { return nil },
			err: ErrMissingHandlerOutputArgument,
		},
		{
			name: "invalid out struct type",
			out: []*config.Parameter{
				{
					Rename: "Test1",
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context) (int, error) { return 0, nil },
			err: ErrWrongOutputArgumentType,
		},
		{
			name: "missing out",
			out: []*config.Parameter{
				{
					Rename: "Test1",
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context) (*struct{}, error) { return nil, nil },
			err: ErrMissingConfigArgument,
		},
		{
			name: "valid out",
			out: []*config.Parameter{
				{
					Rename: "Test1",
					GoType: reflect.TypeOf(int(0)),
				},
			},
			fn:  func(context.Context) (*struct{ Test1 int }, error) { return nil, nil },
			err: nil,
		},
		{
			name: "optional out not ptr",
			out: []*config.Parameter{
				{
					Rename:   "Test1",
					GoType:   reflect.TypeOf(int(0)),
					Optional: true,
				},
			},
			fn:  func(context.Context) (*struct{ Test1 *int }, error) { return nil, nil },
			err: ErrWrongParamTypeFromConfig,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			service := config.Service{
				Input:  make(map[string]*config.Parameter),
				Output: make(map[string]*config.Parameter),
			}

			// fill service with arguments
			if tc.in != nil && len(tc.in) > 0 {
				for i, in := range tc.in {
					service.Input[fmt.Sprintf("%d", i)] = in
				}
			}
			if tc.out != nil && len(tc.out) > 0 {
				for i, out := range tc.out {
					service.Output[fmt.Sprintf("%d", i)] = out
				}
			}

			s := BuildSignature(service)

			err := s.ValidateInput(reflect.TypeOf(tc.fn))
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
				}
				return
			}
			err = s.ValidateOutput(reflect.TypeOf(tc.fn))
			if err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("invalid error\nactual: %v\nexpect: %v", err, tc.err)
				}
				return
			}

			// no error encountered but expected 1
			if tc.err != nil {
				t.Fatalf("expected an error: %v", tc.err)
			}
		})
	}
}
