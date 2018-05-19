package gfw

import (
	"encoding/json"
	"os"
)

// Load builds a struct representation of the
// configuration file located at @path
// The structure checks for most format errors
func Load(path string) (*controller, error) {

	/* (1) Extract data
	---------------------------------------------------------*/
	/* (1) Open file */
	var configFile, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	/* (2) Init receiving dataset */
	var receiver *controller

	/* (3) Decode JSON */
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(receiver)
	if err != nil {
		return nil, err
	}

	return receiver, nil
}
