package gfw

import (
	"git.xdrm.io/xdrm-brackets/gfw/checker"
	"git.xdrm.io/xdrm-brackets/gfw/config"
	"git.xdrm.io/xdrm-brackets/gfw/err"
)

type Server struct {
	config  *config.Controller
	Params  map[string]interface{}
	Checker *checker.TypeRegistry // type check
	err     err.Error
}

type Request struct {
	// corresponds to the list of uri components
	//  featuring in the request URI
	Uri []string

	// portion of the URI that corresponds to the controllerpath
	ControllerUri []string

	// contains all data from URL, GET, and FORM
	Data *requestData
}

type requestData struct {

	// ordered values from the URI
	//  catches all after the controller path
	//
	// points to Request.Data
	Url []*requestParameter

	// uri parameters following the QUERY format
	//
	// points to Request.Data
	Get map[string]*requestParameter

	// form data depending on the Content-Type:
	//  'application/json'                  => key-value pair is parsed as json into the map
	//  'application/x-www-form-urlencoded' => standard parameters as QUERY parameters
	//  'multipart/form-data'               => parse form-data format
	//
	// points to Request.Data
	Form map[string]*requestParameter

	// contains URL+GET+FORM data with prefixes:
	// - FORM: no prefix
	// - URL:  'URL#' followed by the index in Uri
	// - GET:  'GET@' followed by the key in GET
	Set map[string]*requestParameter
}

// requestParameter represents an http request parameter
// that can be of type URL, GET, or FORM (multipart, json, urlencoded)
type requestParameter struct {
	// whether the value has been json-parsed
	// for optimisation purpose, parameters are only parsed
	// if they are required by the current controller
	Parsed bool

	// whether the value is a file
	File bool

	// the actual parameter value
	Value interface{}
}
