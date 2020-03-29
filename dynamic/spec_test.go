package dynamic

import (
	"errors"
	"reflect"
	"testing"
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
		t.Run("case."+string(i), func(t *testing.T) {
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
