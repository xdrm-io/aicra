package meta

import (
	"fmt"
	"git.xdrm.io/go/aicra/driver"
	"os"
	"path/filepath"
)

// format inits the map if not set
func (b *builder) format() {

	if b.Map == nil {
		b.Map = make(map[string]string)
	}

}

// InferFromFolder fills the 'Map' by browsing recursively the
// 'Folder' field
func (b *builder) InferFromFolder(_root string, _driver driver.Driver) {

	// 1. ignore if no Folder
	if len(b.Folder) < 1 {
		return
	}

	// 2. If relative Folder, join to root
	rootpath := filepath.Join(_root, b.Folder)

	// 3. Walk
	filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {

		// ignore dir
		if err != nil || info.IsDir() {
			return nil
		}

		// format path
		path, err = filepath.Rel(rootpath, path)
		if err != nil {
			return nil
		}
		// extract universal path from the driver
		upath := _driver.Path(_root, b.Folder, path)

		// format name
		name := upath
		if name == "ROOT" {
			name = ""
		}
		name = fmt.Sprintf("/%s", name)

		// add to map
		b.Map[name] = upath

		return nil
	})

}
