package request

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

// Parse parameter (json-like) if not already done
func (i *Parameter) Parse() {

	/* (1) Stop if already parsed or nil*/
	if i.Parsed || i.Value == nil {
		return
	}

	/* (2) Try to parse value */
	i.Value = parseParameter(i.Value)

}
