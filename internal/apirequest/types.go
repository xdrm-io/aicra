package apirequest

// Request represents a request by its URI, controller path and data (uri, get, post)
type Request struct {
	// corresponds to the list of uri components
	//  featuring in the request URI
	URI []string

	// controller path (portion of 'Uri')
	Path []string

	// contains all data from URL, GET, and FORM
	Data *DataSet
}

// DataSet represents all data that can be caught:
// - URI (guessed from the URI by removing the controller path)
// - GET (default url data)
// - POST (from json, form-data, url-encoded)
type DataSet struct {

	// ordered values from the URI
	//  catches all after the controller path
	//
	// points to DataSet.Data
	URI []*Parameter

	// uri parameters following the QUERY format
	//
	// points to DataSet.Data
	Get map[string]*Parameter

	// form data depending on the Content-Type:
	//  'application/json'                  => key-value pair is parsed as json into the map
	//  'application/x-www-form-urlencoded' => standard parameters as QUERY parameters
	//  'multipart/form-data'               => parse form-data format
	//
	// points to DataSet.Data
	Form map[string]*Parameter

	// contains URL+GET+FORM data with prefixes:
	// - FORM: no prefix
	// - URL:  'URL#' followed by the index in Uri
	// - GET:  'GET@' followed by the key in GET
	Set map[string]*Parameter
}

// Parameter represents an http request parameter
// that can be of type URL, GET, or FORM (multipart, json, urlencoded)
type Parameter struct {
	// whether the value has been json-parsed
	// for optimisation purpose, parameters are only parsed
	// if they are required by the current controller
	Parsed bool

	// whether the value is a file
	File bool

	// the actual parameter value
	Value interface{}
}
