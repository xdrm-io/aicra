package dynfunc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

type testsignature Signature

// builds a mock service with provided arguments as Input and matched as Output
func (s *testsignature) withArgs(dtypes ...reflect.Type) *testsignature {
	if s.Input == nil {
		s.Input = make(map[string]reflect.Type)
	}
	if s.Output == nil {
		s.Output = make(map[string]reflect.Type)
	}

	for i, dtype := range dtypes {
		name := fmt.Sprintf("P%d", i+1)
		s.Input[name] = dtype
		if dtype.Kind() == reflect.Ptr {
			s.Output[name] = dtype.Elem()
		} else {
			s.Output[name] = dtype
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

	tcases := []struct {
		Name           string
		Spec           *testsignature
		HasContext     bool
		Fn             interface{}
		Input          []interface{}
		ExpectedOutput []interface{}
		ExpectedErr    error
	}{
		{
			Name:           "none required none provided",
			Spec:           (&testsignature{}).withArgs(),
			Fn:             func(context.Context) (*struct{}, error) { return nil, nil },
			HasContext:     false,
			Input:          []interface{}{},
			ExpectedOutput: []interface{}{},
			ExpectedErr:    nil,
		},
		{
			Name: "int proxy (0)",
			Spec: (&testsignature{}).withArgs(reflect.TypeOf(int(0))),
			Fn: func(ctx context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, nil
			},
			HasContext:     false,
			Input:          []interface{}{int(0)},
			ExpectedOutput: []interface{}{int(0)},
			ExpectedErr:    nil,
		},
		{
			Name: "int proxy (11)",
			Spec: (&testsignature{}).withArgs(reflect.TypeOf(int(0))),
			Fn: func(ctx context.Context, in intstruct) (*intstruct, error) {
				return &intstruct{P1: in.P1}, nil
			},
			HasContext:     false,
			Input:          []interface{}{int(11)},
			ExpectedOutput: []interface{}{int(11)},
			ExpectedErr:    nil,
		},
		{
			Name: "*int proxy (nil)",
			Spec: (&testsignature{}).withArgs(reflect.TypeOf(new(int))),
			Fn: func(ctx context.Context, in intptrstruct) (*intptrstruct, error) {
				return &intptrstruct{P1: in.P1}, nil
			},
			HasContext:     false,
			Input:          []interface{}{},
			ExpectedOutput: []interface{}{nil},
			ExpectedErr:    nil,
		},
		{
			Name: "*int proxy (28)",
			Spec: (&testsignature{}).withArgs(reflect.TypeOf(new(int))),
			Fn: func(ctx context.Context, in intptrstruct) (*intstruct, error) {
				return &intstruct{P1: *in.P1}, nil
			},
			HasContext:     false,
			Input:          []interface{}{28},
			ExpectedOutput: []interface{}{28},
			ExpectedErr:    nil,
		},
		{
			Name: "*int proxy (13)",
			Spec: (&testsignature{}).withArgs(reflect.TypeOf(new(int))),
			Fn: func(ctx context.Context, in intptrstruct) (*intstruct, error) {
				return &intstruct{P1: *in.P1}, nil
			},
			HasContext:     false,
			Input:          []interface{}{13},
			ExpectedOutput: []interface{}{13},
			ExpectedErr:    nil,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.Name, func(t *testing.T) {
			t.Parallel()

			var handler = &Handler{
				signature: &Signature{Input: tcase.Spec.Input, Output: tcase.Spec.Output},
				fn:        tcase.Fn,
			}

			// build input
			input := make(map[string]interface{})
			for i, val := range tcase.Input {
				var key = fmt.Sprintf("P%d", i+1)
				input[key] = val
			}

			var output, err = handler.Handle(context.Background(), input)
			if err != tcase.ExpectedErr {
				t.Fatalf("expected api error <%v> got <%v>", tcase.ExpectedErr, err)
			}

			// check output
			for i, expected := range tcase.ExpectedOutput {
				var (
					key         = fmt.Sprintf("P%d", i+1)
					val, exists = output[key]
				)
				if !exists {
					t.Fatalf("missing output[%s]", key)
				}
				if expected != val {
					var (
						expectedt   = reflect.ValueOf(expected)
						valt        = reflect.ValueOf(val)
						expectedNil = !expectedt.IsValid() || expectedt.Kind() == reflect.Ptr && expectedt.IsNil()
						valNil      = !valt.IsValid() || valt.Kind() == reflect.Ptr && valt.IsNil()
					)
					// ignore both nil
					if valNil && expectedNil {
						continue
					}
					t.Fatalf("expected output[%s] to equal %T <%v> got %T <%v>", key, expected, expected, val, val)
				}
			}

		})
	}

}
