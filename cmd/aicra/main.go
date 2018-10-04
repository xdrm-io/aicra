package main

import (
	"fmt"
	"git.xdrm.io/go/aicra/internal/clifmt"
	"git.xdrm.io/go/aicra/internal/config"
	"os"
	"path/filepath"
	"time"
)

var defaultTypeFolder = filepath.Join(os.Getenv("GOPATH"), "src/git.xdrm.io/go/aicra/internal/checker/default/*")

func main() {

	// check argument
	if len(os.Args) < 2 || len(os.Args[1]) < 1 {
		fmt.Printf("missing argument: project path\n")
		return
	}

	// get absolute path from arguments
	root := os.Args[1]
	rootStat, err := os.Stat(root)
	if err != nil || !rootStat.IsDir() {
		fmt.Printf("invalid argument: project path is invalid or not a directory\n")
		return
	}
	if err := os.Chdir(root); err != nil {
		fmt.Printf("invalid argument: cannot chdir to %s\n", root)
		return
	}

	starttime := time.Now()

	/* 1. Load config */
	schema, err := config.Parse("./aicra.json")
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

	/* (2) Compile Default Types */
	if schema.Types.Default {

		clifmt.Title("compile default types")
		files, err := filepath.Glob(defaultTypeFolder)
		if err != nil {
			fmt.Printf("cannot load default types")
		} else {

			for _, file := range files {

				typeName, err := filepath.Rel(filepath.Dir(file), file)
				if err != nil {
					fmt.Printf("cannot load type '%s'\n", typeName)
					continue

				}

				fmt.Printf("   [%s]\n", clifmt.Color(33, typeName))

				// Get useful paths
				source := filepath.Join(file, "main.go")
				build := filepath.Join(schema.Root, ".build/DEFAULT_TYPES", fmt.Sprintf("%s.so", typeName))

				compile(source, build)
			}

		}
	}

	/* (3) Compile Types */
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

	/* (4) Compile controllers */
	if len(schema.Controllers.Map) > 0 {
		clifmt.Title("compile controllers")
		for name, upath := range schema.Controllers.Map {

			fmt.Printf("   [%s]\n", clifmt.Color(33, name))

			// Get useful paths
			source := schema.Driver.Source(schema.Root, schema.Controllers.Folder, upath)
			build := schema.Driver.Build(schema.Root, schema.Controllers.Folder, upath)

			compile(source, build)
		}
	}

	/* (5) Compile middlewares */
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
