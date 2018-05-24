package config

import (
	"encoding/json"
	"os"
	"strings"
)

// Load builds a structured representation of the
// configuration file located at @path
// The struct definition checks for most format errors
func Load(path string) (*Controller, error) {

	/* (1) Extract data
	---------------------------------------------------------*/
	/* (1) Open file */
	var configFile, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	/* (2) Init receiving dataset */
	receiver := &Controller{}

	/* (3) Decode JSON */
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(receiver)
	if err != nil {
		return nil, err
	}

	/* (4) Format result */
	err = receiver.format("/")
	if err != nil {
		return nil, err
	}

	return receiver, nil

}

// IsMethodAvailable returns whether a given
// method is available (case insensitive)
func IsMethodAvailable(method string) bool {
	for _, m := range AvailableMethods {
		if strings.ToUpper(method) == m {
			return true
		}
	}

	return false
}

// HasMethod returns whether the controller has a given
// method (case insensitive)
func (c Controller) HasMethod(method string) bool {
	method = strings.ToUpper(method)

	switch method {

	case "GET":
		return c.GET != nil
	case "POST":
		return c.POST != nil
	case "PUT":
		return c.PUT != nil
	case "DELETE":
		return c.DELETE != nil
	default:
		return false

	}

}
