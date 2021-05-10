package dynfunc

import (
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
			Fn:    func() {},
			FnCtx: func(api.Ctx) {},
			Err:   nil,
		},
		{
			Name:  "no input 1 given",
			Input: map[string]reflect.Type{},
			Fn:    func(int) {},
			FnCtx: func(api.Ctx, int) {},
			Err:   errUnexpectedInput,
		},
		{
			Name:  "no input 2 given",
			Input: map[string]reflect.Type{},
			Fn:    func(int, string) {},
			FnCtx: func(api.Ctx, int, string) {},
			Err:   errUnexpectedInput,
		},
		{
			Name: "1 input 0 given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func() {},
			FnCtx: func(api.Ctx) {},
			Err:   errMissingHandlerArgumentParam,
		},
		{
			Name: "1 input non-struct given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(int) {},
			FnCtx: func(api.Ctx, int) {},
			Err:   errMissingParamArgument,
		},
		{
			Name: "unexported input",
			Input: map[string]reflect.Type{
				"test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(struct{}) {},
			FnCtx: func(api.Ctx, struct{}) {},
			Err:   errUnexportedName,
		},
		{
			Name: "1 input empty struct given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(struct{}) {},
			FnCtx: func(api.Ctx, struct{}) {},
			Err:   errMissingParamFromConfig,
		},
		{
			Name: "1 input invalid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(struct{ Test1 string }) {},
			FnCtx: func(api.Ctx, struct{ Test1 string }) {},
			Err:   errWrongParamTypeFromConfig,
		},
		{
			Name: "1 input valid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:    func(struct{ Test1 int }) {},
			FnCtx: func(api.Ctx, struct{ Test1 int }) {},
			Err:   nil,
		},
		{
			Name: "1 input ptr empty struct given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(struct{}) {},
			FnCtx: func(api.Ctx, struct{}) {},
			Err:   errMissingParamFromConfig,
		},
		{
			Name: "1 input ptr invalid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(struct{ Test1 string }) {},
			FnCtx: func(api.Ctx, struct{ Test1 string }) {},
			Err:   errWrongParamTypeFromConfig,
		},
		{
			Name: "1 input ptr invalid ptr type given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(struct{ Test1 *string }) {},
			FnCtx: func(api.Ctx, struct{ Test1 *string }) {},
			Err:   errWrongParamTypeFromConfig,
		},
		{
			Name: "1 input ptr valid given",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(int)),
			},
			Fn:    func(struct{ Test1 *int }) {},
			FnCtx: func(api.Ctx, struct{ Test1 *int }) {},
			Err:   nil,
		},
		{
			Name: "1 valid string",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(string("")),
			},
			Fn:    func(struct{ Test1 string }) {},
			FnCtx: func(api.Ctx, struct{ Test1 string }) {},
			Err:   nil,
		},
		{
			Name: "1 valid uint",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(uint(0)),
			},
			Fn:    func(struct{ Test1 uint }) {},
			FnCtx: func(api.Ctx, struct{ Test1 uint }) {},
			Err:   nil,
		},
		{
			Name: "1 valid float64",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(float64(0)),
			},
			Fn:    func(struct{ Test1 float64 }) {},
			FnCtx: func(api.Ctx, struct{ Test1 float64 }) {},
			Err:   nil,
		},
		{
			Name: "1 valid []byte",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf([]byte("")),
			},
			Fn:    func(struct{ Test1 []byte }) {},
			FnCtx: func(api.Ctx, struct{ Test1 []byte }) {},
			Err:   nil,
		},
		{
			Name: "1 valid []rune",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf([]rune("")),
			},
			Fn:    func(struct{ Test1 []rune }) {},
			FnCtx: func(api.Ctx, struct{ Test1 []rune }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *string",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(string)),
			},
			Fn:    func(struct{ Test1 *string }) {},
			FnCtx: func(api.Ctx, struct{ Test1 *string }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *uint",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(uint)),
			},
			Fn:    func(struct{ Test1 *uint }) {},
			FnCtx: func(api.Ctx, struct{ Test1 *uint }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *float64",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new(float64)),
			},
			Fn:    func(struct{ Test1 *float64 }) {},
			FnCtx: func(api.Ctx, struct{ Test1 *float64 }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *[]byte",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new([]byte)),
			},
			Fn:    func(struct{ Test1 *[]byte }) {},
			FnCtx: func(api.Ctx, struct{ Test1 *[]byte }) {},
			Err:   nil,
		},
		{
			Name: "1 valid *[]rune",
			Input: map[string]reflect.Type{
				"Test1": reflect.TypeOf(new([]rune)),
			},
			Fn:    func(struct{ Test1 *[]rune }) {},
			FnCtx: func(api.Ctx, struct{ Test1 *[]rune }) {},
			Err:   nil,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.Name, func(t *testing.T) {
			t.Parallel()

			// mock spec
			s := signature{
				Input:  tcase.Input,
				Output: nil,
			}

			t.Run("with-context", func(t *testing.T) {
				err := s.checkInput(reflect.TypeOf(tcase.FnCtx), 1)
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
			t.Run("without-context", func(t *testing.T) {
				err := s.checkInput(reflect.TypeOf(tcase.Fn), 0)
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
			Fn:     func() {},
			Err:    errMissingHandlerOutput,
		},
		// no input -> with last type not api.Err
		{
			Output: map[string]reflect.Type{},
			Fn:     func() bool { return true },
			Err:    errMissingHandlerErrorOutput,
		},
		// no input -> with api.Err
		{
			Output: map[string]reflect.Type{},
			Fn:     func() api.Err { return api.ErrSuccess },
			Err:    nil,
		},
		// func can have output if not specified
		{
			Output: map[string]reflect.Type{},
			Fn:     func() (*struct{}, api.Err) { return nil, api.ErrSuccess },
			Err:    nil,
		},
		// missing output struct in func
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() api.Err { return api.ErrSuccess },
			Err: errMissingParamOutput,
		},
		// output not a pointer
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (int, api.Err) { return 0, api.ErrSuccess },
			Err: errMissingParamOutput,
		},
		// output not a pointer to struct
		{
			Output: map[string]reflect.Type{
				"Test1": reflect.TypeOf(int(0)),
			},
			Fn:  func() (*int, api.Err) { return nil, api.ErrSuccess },
			Err: errMissingParamOutput,
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
			Err: errMissingParamFromConfig,
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
			s := signature{
				Input:  nil,
				Output: tcase.Output,
			}

			err := s.checkOutput(reflect.TypeOf(tcase.Fn))
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
