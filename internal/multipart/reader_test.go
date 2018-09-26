package multipart

import (
	"bytes"
	"testing"
)

func TestSimple(t *testing.T) {
	test := struct {
		Input    []byte
		Boundary string
	}{
		Input: []byte(`--BoUnDaRy
Content-Disposition: form-data; name="somevar"

google.com
--BoUnDaRy
Content-Disposition: form-data; name="somefile"; filename="somefilename.pdf"
Content-Type: application/pdf

facebook.com
--BoUnDaRy--`),
		Boundary: "BoUnDaRy",
	}

	mpr, err := NewReader(bytes.NewReader(test.Input), test.Boundary)

	if err != nil {
		t.Fatalf("Unexpected error <%s>", err)
	}

	if err = mpr.Parse(); err != nil {
		t.Fatalf("Unexpected error <%s>", err)
	}

	// 1. Check var
	somevar := mpr.Get("somevar")
	if somevar == nil {
		t.Fatalf("Expected data {%s} to exist", "somevar")
	}
	if somevar.ContentType != "raw" {
		t.Fatalf("Expected ContentType to be {raw}, got {%s}", somevar.ContentType)
	}

	if string(somevar.Data) != "google.com" {
		t.Fatalf("Expected data to be {%s}, got {%s}", "google.com", somevar.Data)
	}

	// 2. Check file
	somefile := mpr.Get("somefile")
	if somefile == nil {
		t.Fatalf("Expected data {%s} to exist", "somefile")
	}
	if somefile.ContentType != "application/pdf" {
		t.Fatalf("Expected ContentType to be {application/pdf}, got {%s}", somevar.ContentType)
	}

	if string(somefile.Data) != "facebook.com" {
		t.Fatalf("Expected data to be {%s}, got {%s}", "facebook.com", somefile.Data)
	}

	filename := somefile.GetHeader("filename")
	if len(filename) < 1 {
		t.Fatalf("Expected data to have header 'filename'")
	}
	if filename != "somefilename.pdf" {
		t.Fatalf("Expected filename to be {%s}, got {%s}", "somefilename.pdf", filename)
	}

}

func TestSimpleWithCRLF(t *testing.T) {

	type tcase struct {
		Input    []byte
		Boundary string
	}
	_test := tcase{
		Input: []byte(`--BoUnDaRy
Content-Disposition: form-data; name="somevar"

google.com
--BoUnDaRy
Content-Disposition: form-data; name="somefile"; filename="somefilename.pdf"
Content-Type: application/pdf

facebook.com
--BoUnDaRy--`),
		Boundary: "BoUnDaRy",
	}

	test := tcase{
		Input:    make([]byte, 0),
		Boundary: _test.Boundary,
	}

	// replace all \n with \r\n
	for _, char := range _test.Input {
		if char == '\n' {
			test.Input = append(test.Input, []byte("\r\n")...)
			continue
		}

		test.Input = append(test.Input, char)
	}

	mpr, err := NewReader(bytes.NewReader(test.Input), test.Boundary)

	if err != nil {
		t.Fatalf("Unexpected error <%s>", err)
	}

	if err = mpr.Parse(); err != nil {
		t.Fatalf("Unexpected error <%s>", err)
	}

	// 1. Check var
	somevar := mpr.Get("somevar")
	if somevar == nil {
		t.Fatalf("Expected data {%s} to exist", "somevar")
	}
	if somevar.ContentType != "raw" {
		t.Fatalf("Expected ContentType to be {raw}, got {%s}", somevar.ContentType)
	}

	if string(somevar.Data) != "google.com" {
		t.Fatalf("Expected data to be {%s}, got {%s}", "google.com", somevar.Data)
	}

	// 2. Check file
	somefile := mpr.Get("somefile")
	if somefile == nil {
		t.Fatalf("Expected data {%s} to exist", "somefile")
	}
	if somefile.ContentType != "application/pdf" {
		t.Fatalf("Expected ContentType to be {application/pdf}, got {%s}", somevar.ContentType)
	}

	if string(somefile.Data) != "facebook.com" {
		t.Fatalf("Expected data to be {%s}, got {%s}", "facebook.com", somefile.Data)
	}

	filename := somefile.GetHeader("filename")
	if len(filename) < 1 {
		t.Fatalf("Expected data to have header 'filename'")
	}
	if filename != "somefilename.pdf" {
		t.Fatalf("Expected filename to be {%s}, got {%s}", "somefilename.pdf", filename)
	}

}

func TestNoName(t *testing.T) {
	tests := []struct {
		Input    []byte
		Boundary string
		Length   int
	}{
		{
			Input:    []byte("--BoUnDaRy\nContent-Disposition: form-data; xname=\"somevar\"\n\ngoogle.com\n--BoUnDaRy--"),
			Boundary: "BoUnDaRy",
		},
		{
			Input:    []byte("--BoUnDaRy\nContent-Disposition: form-data; name=\"\"\n\ngoogle.com\n--BoUnDaRy--"),
			Boundary: "BoUnDaRy",
		},
		{
			Input:    []byte("--BoUnDaRy\nContent-Disposition: form-data; name=\n\ngoogle.com\n--BoUnDaRy--"),
			Boundary: "BoUnDaRy",
		},
		{
			Input:    []byte("--BoUnDaRy\nContent-Disposition: form-data; name\n\ngoogle.com\n--BoUnDaRy--"),
			Boundary: "BoUnDaRy",
		},
	}

	for i, test := range tests {

		mpr, err := NewReader(bytes.NewReader(test.Input), test.Boundary)

		if err != nil {
			t.Errorf("(%d) Unexpected error <%s>", i, err)
			continue
		}

		if err = mpr.Parse(); err != ErrMissingDataName {
			t.Errorf("(%d) Expected the error <%s>, got <%s>", i, ErrMissingDataName, err)
			continue
		}

	}

}

func TestNoHeader(t *testing.T) {
	tests := []struct {
		Input    []byte
		Boundary string
		Length   int
	}{
		{
			Input:    []byte("--BoUnDaRy\n\ngoogle.com\n--BoUnDaRy--"),
			Boundary: "BoUnDaRy",
		},
		{
			Input:    []byte("--BoUnDaRy\nContent-Disposition: false;\n\ngoogle.com\n--BoUnDaRy--"),
			Boundary: "BoUnDaRy",
		},
		{
			Input:    []byte("--BoUnDaRy\nContent-Disposition: form-data;\n\ngoogle.com\n--BoUnDaRy--"),
			Boundary: "BoUnDaRy",
		},
	}

	for i, test := range tests {

		mpr, err := NewReader(bytes.NewReader(test.Input), test.Boundary)

		if err != nil {
			t.Errorf("(%d) Unexpected error <%s>", i, err)
			continue
		}

		if err = mpr.Parse(); err != ErrNoHeader {
			t.Errorf("(%d) Expected the error <%s>, got <%s>", i, ErrNoHeader, err)
			continue
		}

	}

}

func TestNameConflict(t *testing.T) {
	test := struct {
		Input    []byte
		Boundary string
		Length   int
	}{
		Input: []byte(`--BoUnDaRy
Content-Disposition: form-data; name="var1"

google.com
--BoUnDaRy
Content-Disposition: form-data; name="var1"

facebook.com
--BoUnDaRy--`),
		Boundary: "BoUnDaRy",
	}

	mpr, err := NewReader(bytes.NewReader(test.Input), test.Boundary)

	if err != nil {
		t.Fatalf("Unexpected error <%s>", err)
	}

	if err = mpr.Parse(); err != ErrDataNameConflict {
		t.Fatalf("Expected the error <%s>, got <%s>", ErrDataNameConflict, err)
	}

}

func TestGetterNil(t *testing.T) {
	test := struct {
		Input    []byte
		Boundary string
		Length   int
	}{
		Input: []byte(`--BoUnDaRy
Content-Disposition: form-data; name="var1"

google.com
--BoUnDaRy
Content-Disposition: form-data; name="var2"

facebook.com
--BoUnDaRy--`),
		Boundary: "BoUnDaRy",
	}

	mpr, err := NewReader(bytes.NewReader(test.Input), test.Boundary)

	if err != nil {
		t.Fatalf("Unexpected error <%s>", err)
	}

	if err = mpr.Parse(); err != nil {
		t.Fatalf("Unexpected error <%s>", err)
	}

	if mpr.Get("unknown_key") != nil {
		t.Fatalf("Expected 'unknown_key' not to exist, got {%v}", mpr.Get("unknown_key"))
	}

}
