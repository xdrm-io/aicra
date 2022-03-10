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
