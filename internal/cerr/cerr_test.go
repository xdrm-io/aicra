package cerr

import (
	"errors"
	"fmt"
	"testing"
)

func TestConstError(t *testing.T) {
	const cerr1 = Error("some-string")
	const cerr2 = Error("some-other-string")
	const cerr3 = Error("some-string") // same const value as @cerr1

	if cerr1.Error() == cerr2.Error() {
		t.Errorf("cerr1 should not be equal to cerr2 ('%s', '%s')", cerr1.Error(), cerr2.Error())
	}
	if cerr2.Error() == cerr3.Error() {
		t.Errorf("cerr2 should not be equal to cerr3 ('%s', '%s')", cerr2.Error(), cerr3.Error())
	}
	if cerr1.Error() != cerr3.Error() {
		t.Errorf("cerr1 should be equal to cerr3 ('%s', '%s')", cerr1.Error(), cerr3.Error())
	}
}

func TestWrappedConstError(t *testing.T) {
	const parent = Error("file error")

	const readErrorConst = Error("cannot read file")
	var wrappedReadError = parent.Wrap(readErrorConst)

	expectedWrappedReadError := fmt.Sprintf("%s: %s", parent.Error(), readErrorConst.Error())
	if wrappedReadError.Error() != expectedWrappedReadError {
		t.Errorf("expected '%s' (got '%s')", wrappedReadError.Error(), expectedWrappedReadError)
	}
}
func TestWrappedStandardError(t *testing.T) {
	const parent = Error("file error")

	var writeErrorStandard error = errors.New("cannot write file")
	var wrappedWriteError = parent.Wrap(writeErrorStandard)

	expectedWrappedWriteError := fmt.Sprintf("%s: %s", parent.Error(), writeErrorStandard.Error())
	if wrappedWriteError.Error() != expectedWrappedWriteError {
		t.Errorf("expected '%s' (got '%s')", wrappedWriteError.Error(), expectedWrappedWriteError)
	}
}
func TestWrappedStringError(t *testing.T) {
	const parent = Error("file error")

	var closeErrorString string = "cannot close file"
	var wrappedCloseError = parent.WrapString(closeErrorString)

	expectedWrappedCloseError := fmt.Sprintf("%s: %s", parent.Error(), closeErrorString)
	if wrappedCloseError.Error() != expectedWrappedCloseError {
		t.Errorf("expected '%s' (got '%s')", wrappedCloseError.Error(), expectedWrappedCloseError)
	}
}
