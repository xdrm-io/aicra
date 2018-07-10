package apirequest

// Parse parameter (json-like) if not already done
func (i *Parameter) Parse() {

	/* (1) Stop if already parsed or nil*/
	if i.Parsed || i.Value == nil {
		return
	}

	/* (2) Try to parse value */
	i.Value = parseParameter(i.Value)

}
