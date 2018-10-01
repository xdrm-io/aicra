package meta

import (
	"encoding/json"
	"errors"
	"git.xdrm.io/go/aicra/driver"
	"os"
	"path/filepath"
	"strings"
)

// Parse extracts a Meta from a json config file (aicra.json)
func Parse(_path string) (*Schema, error) {

	/* 1. ppen file */
	file, err := os.Open(_path)
	if err != nil {
		return nil, errors.New("cannot open file")
	}
	defer file.Close()

	/* 2. Init receiver dataset */
	receiver := &Schema{}

	/* 3. Decode json */
	decoder := json.NewDecoder(file)
	err = decoder.Decode(receiver)
	if err != nil {
		return nil, err
	}

	/* 4. Error on invalid driver */
	receiver.DriverName = strings.ToLower(receiver.DriverName)
	switch receiver.DriverName {
	case "generic":
		receiver.Driver = &driver.Generic{}
	case "plugin":
		receiver.Driver = &driver.Plugin{}
	case "import":
		receiver.Driver = &driver.Import{}

	default:
		return nil, errors.New("invalid driver; choose from 'generic', 'plugin', and 'import'")
	}

	/* 5. Fail if type map is set */
	if receiver.Types.Map != nil {
		return nil, errors.New("types must not feature the 'map'")
	}

	/* 6. Fail on absolute folders */
	if len(receiver.Types.Folder) > 0 && filepath.IsAbs(receiver.Types.Folder) {
		return nil, errors.New("types folder must be relative to root")
	}
	if len(receiver.Controllers.Folder) > 0 && filepath.IsAbs(receiver.Controllers.Folder) {
		return nil, errors.New("controllers folder must be relative to root")
	}
	if len(receiver.Middlewares.Folder) > 0 && filepath.IsAbs(receiver.Middlewares.Folder) {
		return nil, errors.New("middlewares folder must be relative to root")
	}

	/* 7. Format result (default values, etc) */
	receiver.setDefaults()

	return receiver, nil

}

// setDefaults sets defaults values and checks for missing data
func (m *Schema) setDefaults() {

	// 1. extract absolute root folder
	absroot, err := filepath.Abs(m.Root)
	if err == nil {
		m.Root = absroot
	}

	// 2. host
	if len(m.Host) < 1 {
		m.Host = Default.Host
	}

	// 3. port
	if m.Port == 0 {
		m.Port = Default.Port
	}

	// 4. Empty builders
	if m.Types == nil {
		m.Types = Default.Types
	}
	if m.Controllers == nil {
		m.Controllers = Default.Controllers
	}
	if m.Middlewares == nil {
		m.Middlewares = Default.Middlewares
	}

	// 5. Init map if not set
	m.Types.format()
	m.Controllers.format()
	m.Middlewares.format()

	// 6. Infer Maps from Folders
	m.Types.InferFromFolder(m.Root, m.Driver)
	m.Controllers.InferFromFolder(m.Root, m.Driver)
	m.Middlewares.InferFromFolder(m.Root, m.Driver)

}