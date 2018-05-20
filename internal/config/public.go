package config

import (
	"encoding/json"
	"os"
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

	return receiver, nil

}
