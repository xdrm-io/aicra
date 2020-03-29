package dynamic

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"git.xdrm.io/go/aicra/api"
)

func TestInputCheck(t *testing.T) {
	tcases := []struct {
		Input map[string]reflect.Type
		Fn    interface{}
		Err   error
	}{
		// no input
		{
			Input: map[string]reflect.Type{},
			Fn:    func() {},
			Err:   nil,
		},
		// func can have any arguments if not specified
		{
			Input: map[string]reflect.Type{},
			Fn:    func(int, string) {},
			Err:   nil,
		},
		// missing input struct in func
		{
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() {},
			Err: ErrMissingHandlerArgumentParam,
		},
		// input not a struct
		{
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func(int) {},
			Err: ErrMissingParamArgument,
		},
		// unexported param name
		{
			Input: map[string]reflect.Type{
				"test1": reflect.TypeOf(int(0)),
			},
			Fn:  func(struct{}) {},
			Err: ErrUnexportedName,
		},
		// input field missing
		{
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func(struct{}) {},
			Err: ErrMissingParamFromConfig,
		},
		// input field invalid type
		{
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func(struct{ Test1 string }) {},
			Err: ErrWrongParamTypeFromConfig,
		},
		// input field valid type
		{
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func(struct{ Test1 int }) {},
			Err: nil,
		},
	}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("case.%d", i), func(t *testing.T) {
			// mock spec
			s := spec{
				Input:  tcase.Input,
				Output: nil,
			}

			err := s.checkInput(reflect.ValueOf(tcase.Fn))
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
		// no input -> missing api.Error
		{
			Output: map[string]reflect.Type{},
			Fn:     func() {},
			Err:    ErrMissingHandlerOutput,
		},
		// no input -> with api.Error
		{
			Output: map[string]reflect.Type{},
			Fn:     func() api.Error { return api.ErrorSuccess },
			Err:    nil,
		},
		// func can have output if not specified
		{
			Output: map[string]reflect.Type{},
			Fn:     func() (*struct{}, api.Error) { return nil, api.ErrorSuccess },
			Err:    nil,
		},
		// missing output struct in func
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() api.Error { return api.ErrorSuccess },
			Err: ErrMissingParamOutput,
		},
		// output not a pointer
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (int, api.Error) { return 0, api.ErrorSuccess },
			Err: ErrMissingParamOutput,
		},
		// output not a pointer to struct
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*int, api.Error) { return nil, api.ErrorSuccess },
			Err: ErrMissingParamOutput,
		},
		// unexported param name
		{
			Output: map[string]reflect.Type{
				"test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{}, api.Error) { return nil, api.ErrorSuccess },
			Err: ErrUnexportedName,
		},
		// output field missing
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{}, api.Error) { return nil, api.ErrorSuccess },
			Err: ErrMissingParamFromConfig,
		},
		// output field invalid type
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{ Test1 string }, api.Error) { return nil, api.ErrorSuccess },
			Err: ErrWrongParamTypeFromConfig,
		},
		// output field valid type
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*struct{ Test1 int }, api.Error) { return nil, api.ErrorSuccess },
			Err: nil,
		},
	}

	for i, tcase := range tcases {
		t.Run(fmt.Sprintf("case.%d", i), func(t *testing.T) {
			// mock spec
			s := spec{
				Input:  nil,
				Output: tcase.Output,
			}

			err := s.checkOutput(reflect.ValueOf(tcase.Fn))
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
