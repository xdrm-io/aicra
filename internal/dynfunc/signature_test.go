package dynfunc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"git.xdrm.io/go/aicra/api"
)

func TestInputCheck(t *testing.T) {
	tcases := []struct {
		Name  string
		Input map[string]reflect.Type
		Fn    interface{}
		FnCtx interface{}
		Err   error
	}{
		{
			Name:  "no input 0 given",
			Input: map[string]reflect.Type{},
			Fn:    func(context.Context) {},
			FnCtx: func(context.Context) {},
			Err:   nil,
		},
		{
			Name:  "no input 1 given",
			Input: map[string]reflect.Type{},
			Fn:    func(context.Context, int) {},
			FnCtx: func(context.Context, int) {},
			Err:   errUnexpectedInput,
		},
		{
			Name:  "no input 2 given",
			Input: map[string]reflect.Type{},
			Fn:    func(context.Context, int, string) {},
			FnCtx: func(context.Context, int, string) {},
			Err:   errUnexpectedInput,
		},
		{
			Name: "1 input 0 given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(context.Context) {},
			FnCtx: func(context.Context) {},
			Err:   errMissingHandlerInputArgument,
		},
		{
			Name: "1 input non-struct given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(context.Context, int) {},
			FnCtx: func(context.Context, int) {},
			Err:   errMissingParamArgument,
		},
		{
			Name: "unexported input",
			Input: map[string]reflect.Type{
				"test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(context.Context, struct{}) {},
			FnCtx: func(context.Context, struct{}) {},
			Err:   errUnexportedName,
		},
		{
			Name: "1 input empty struct given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(context.Context, struct{}) {},
			FnCtx: func(context.Context, struct{}) {},
			Err:   errMissingConfigArgument,
		},
		{
			Name: "1 input invalid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(context.Context, struct{ Test1 string }) {},
			FnCtx: func(context.Context, struct{ Test1 string }) {},
			Err:   errWrongParamTypeFromConfig,
		},
		{
			Name: "1 input valid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(context.Context, struct{ Test1 int }) {},
			FnCtx: func(context.Context, struct{ Test1 int }) {},
			Err:   nil,
		},
		{
			Name: "1 input ptr empty struct given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(context.Context, struct{}) {},
			FnCtx: func(context.Context, struct{}) {},
			Err:   errMissingConfigArgument,
		},
		{
			Name: "1 input ptr invalid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(context.Context, struct{ Test1 string }) {},
			FnCtx: func(context.Context, struct{ Test1 string }) {},
			Err:   errWrongParamTypeFromConfig,
		},
		{
			Name: "1 input ptr invalid ptr type given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(context.Context, struct{ Test1 *string }) {},
			FnCtx: func(context.Context, struct{ Test1 *string }) {},
			Err:   errWrongParamTypeFromConfig,
		},
		{
			Name: "1 input ptr valid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(context.Context, struct{ Test1 *int }) {},
			FnCtx: func(context.Context, struct{ Test1 *int }) {},
			Err:   nil,
		},
		{
			Name: "1 valid string",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(string("")),
			},
			Fn:    func(context.Context, struct{ Test1 string }) {},
			FnCtx: func(context.Context, struct{ Test1 string }) {},
			Err:   nil,
		},
		{
			Name: "1 valid uint",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(uint(0)),
			},
			Fn:    func(context.Context, struct{ Test1 uint }) {},
			FnCtx: func(context.Context, struct{ Test1 uint }) {},
			Err:   nil,
		},
		{
			Name: "1 valid float64",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(float64(0)),
			},
			Fn:    func(context.Context, struct{ Test1 float64 }) {},
			FnCtx: func(context.Context, struct{ Test1 float64 }) {},
			Err:   nil,
		},
		{
			Name: "1 valid []byte",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf([]byte("")),
			},
			Fn:    func(context.Context, struct{ Test1 []byte }) {},
			FnCtx: func(context.Context, struct{ Test1 []byte }) {},
			Err:   nil,
		},
		{
			Name: "1 valid []rune",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf([]rune("")),
			},
			Fn:    func(context.Context, struct{ Test1 []rune }) {},
			FnCtx: func(context.Context, struct{ Test1 []rune }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *string",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(string)),
			},
			Fn:    func(context.Context, struct{ Test1 *string }) {},
			FnCtx: func(context.Context, struct{ Test1 *string }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *uint",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(uint)),
			},
			Fn:    func(context.Context, struct{ Test1 *uint }) {},
			FnCtx: func(context.Context, struct{ Test1 *uint }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *float64",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(float64)),
			},
			Fn:    func(context.Context, struct{ Test1 *float64 }) {},
			FnCtx: func(context.Context, struct{ Test1 *float64 }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *[]byte",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new([]byte)),
			},
			Fn:    func(context.Context, struct{ Test1 *[]byte }) {},
			FnCtx: func(context.Context, struct{ Test1 *[]byte }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *[]rune",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new([]rune)),
			},
			Fn:    func(context.Context, struct{ Test1 *[]rune }) {},
			FnCtx: func(context.Context, struct{ Test1 *[]rune }) {},
			Err:   nil,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.Name, func(t *testing.T) {
			t.Parallel()

			// mock spec
			s := Signature{
				Input:  tcase.Input,
				Output: nil,
			}

			err := s.ValidateInput(reflect.TypeOf(tcase.FnCtx))
			if err == nil && tcase.Err != nil {
				t.Errorf("expected an error: '%s'", tcase.Err.Error())
				t.FailNow()
			}
			if err != nil && tcase.Err == nil {
				t.Errorf("unexpected error: '%s'", err.Error())
				t.FailNow()
			}

			if err != nil && tcase.Err != nil {
				if !errors.Is(err, tcase.Err) {
					t.Errorf("expected the error <%s> got <%s>", tcase.Err, err)
					t.FailNow()
				}
			}
		})
	}
}

func TestOutputCheck(t *testing.T) {
	tcases := []struct {
		Output map[string]reflect.Type
		Fn     interface{}
		Err    error
	}{
		// no input -> missing api.Err
		{
			Output: map[string]reflect.Type{},
			Fn:     func(context.Context) {},
			Err:    errMissingHandlerOutputArgument,
		},
		// no input -> with last type not api.Err
		{
			Output: map[string]reflect.Type{},
			Fn:     func(context.Context) bool { return true },
			Err:    errMissingHandlerErrorArgument,
		},
		// no input -> with api.Err
		{
			Output: map[string]reflect.Type{},
			Fn:     func(context.Context) api.Err { return api.ErrSuccess },
			Err:    nil,
		},
		// no input -> missing context.Context
		{
			Output: map[string]reflect.Type{},
			Fn:     func(context.Context) api.Err { return api.ErrSuccess },
			Err:    errMissingHandlerContextArgument,
		},
		// no input -> invlaid context.Context type
		{
			Output: map[string]reflect.Type{},
			Fn:     func(context.Context, int) api.Err { return api.ErrSuccess },
			Err:    errMissingHandlerContextArgument,
		},
		// func can have output if not specified
		{
			Output: map[string]reflect.Type{},
			Fn:     func(context.Context) (*struct{}, api.Err) { return nil, api.ErrSuccess },
			Err:    nil,
		},
		// missing output struct in func
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() api.Err { return api.ErrSuccess },
			Err: errWrongOutputArgumentType,
		},
		// output not a pointer
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (int, api.Err) { return 0, api.ErrSuccess },
			Err: errWrongOutputArgumentType,
		},
		// output not a pointer to struct
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*int, api.Err) { return nil, api.ErrSuccess },
			Err: errWrongOutputArgumentType,
		},
		// unexported param name
		{
			Output: map[string]reflect.Type{
				"test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{}, api.Err) { return nil, api.ErrSuccess },
			Err: errUnexportedName,
		},
		// output field missing
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{}, api.Err) { return nil, api.ErrSuccess },
			Err: errMissingConfigArgument,
		},
		// output field invalid type
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{ Test1 string }, api.Err) { return nil, api.ErrSuccess },
			Err: errWrongParamTypeFromConfig,
		},
		// output field valid type
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{ Test1 int }, api.Err) { return nil, api.ErrSuccess },
			Err: nil,
		},
		// ignore type check on nil type
		{
			Output: map[string]reflect.Type{
				"Test1": nil,
			},
			Fn:  func() (*struct{ Test1 int }, api.Err) { return nil, api.ErrSuccess },
			Err: nil,
		},
	}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("case.%d", i), func(t *testing.T) {
			t.Parallel()

			// mock spec
			s := Signature{
				Input:  nil,
				Output: tcase.Output,
			}

			err := s.ValidateOutput(reflect.TypeOf(tcase.Fn))
			if err == nil && tcase.Err != nil {
				t.Errorf("expected an error: '%s'", tcase.Err.Error())
				t.FailNow()
			}
			if err != nil && tcase.Err == nil {
				t.Errorf("unexpected error: '%s'", err.Error())
				t.FailNow()
			}

			if err != nil && tcase.Err != nil {
				if !errors.Is(err, tcase.Err) {
					t.Errorf("expected the error <%s> got <%s>", tcase.Err, err)
					t.FailNow()
				}
			}
		})
	}
}
