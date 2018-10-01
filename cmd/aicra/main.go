package main

import (
	"fmt"
	"git.xdrm.io/go/aicra/internal/clifmt"
	"git.xdrm.io/go/aicra/internal/meta"
	"os"
	"path/filepath"
	"time"
)

func main() {

	starttime := time.Now()

	/* 1. Load config */
	schema, err := meta.Parse("./aicra.json")
	if err != nil {
		fmt.Printf("aicra.json: %s\n", err)
		return
	}

	/* 2. End if nothing to compile */
	if !schema.Driver.Compiled() {
		fmt.Printf("\n[ %s | %s ] nothing to compile\n",
			clifmt.Color(32, "finished"),
			time.Now().Sub(starttime),
		)
		return
	}

	/* Compile
	---------------------------------------------------------*/
	/* (1) Create build output dir */
	buildPath := filepath.Join(schema.Root, ".build")
	err = os.MkdirAll(buildPath, os.ModePerm)
	if err != nil {
		fmt.Printf("%s the directory %s cannot be created, check permissions.", clifmt.Warn(), clifmt.Color(33, buildPath))
		return
	}

	/* (2) Compile Types */
	if len(schema.Types.Map) > 0 {
		clifmt.Title("compile types")
		for name, upath := range schema.Types.Map {

			fmt.Printf("   [%s]\n", clifmt.Color(33, name))

			// Get useful paths
			source := schema.Driver.Source(schema.Root, schema.Types.Folder, upath)
			build := schema.Driver.Build(schema.Root, schema.Types.Folder, upath)

			compile(source, build)
		}
	}

	/* (3) Compile controllers */
	if len(schema.Controllers.Map) > 0 {
		clifmt.Title("compile controllers")
		for name, upath := range schema.Controllers.Map {

			fmt.Printf("   [%s]\n", clifmt.Color(33, name))

			// Get useful paths
			source := schema.Driver.Source(schema.Root, schema.Controllers.Folder, upath)
			build := schema.Driver.Build(schema.Root, schema.Controllers.Folder, upath)

			compile(source, build)
			fmt.Printf("\n")
		}
	}

	/* (4) Compile middlewares */
	if len(schema.Middlewares.Map) > 0 {
		clifmt.Title("compile middlewares")
		for name, upath := range schema.Middlewares.Map {

			fmt.Printf("   [%s]\n", clifmt.Color(33, name))

			// Get useful paths
			source := schema.Driver.Source(schema.Root, schema.Middlewares.Folder, upath)
			build := schema.Driver.Build(schema.Root, schema.Middlewares.Folder, upath)

			compile(source, build)
		}
	}

	/* (6) finished */
	fmt.Printf("\n[ %s | %s ] files are located inside the %s directory inside the project folder\n",
		clifmt.Color(32, "finished"),
		time.Now().Sub(starttime),
		clifmt.Color(33, ".build"),
	)

}
