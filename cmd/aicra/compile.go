package main

import (
	"fmt"
	"git.xdrm.io/go/aicra/clifmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Builds types as plugins (.so)
// from the sources in the @in folder
// recursively and generate .so files
// into the @out folder with the same structure
func compile(in string, out string) error {

	/* (1) Create build folder */
	clifmt.Align("    . create output folder")
	err := os.MkdirAll(out, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Printf("ok\n")

	/* (2) List recursively */
	types := []string{}
	err = filepath.Walk(in, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, "/main.go") {
			types = append(types, filepath.Base(filepath.Dir(path)))
		}
		return nil
	})

	if err != nil {
		return err
	}

	/* (3) Print files */
	for _, name := range types {

		// 1. process output file name
		infile := filepath.Join(in, name, "main.go")
		outfile := filepath.Join(out, fmt.Sprintf("%s.so", name))

		clifmt.Align(fmt.Sprintf("    . compile %s", clifmt.Color(33, name)))

		// 2. compile
		stdout, err := exec.Command("go",
			"build", "-buildmode=plugin",
			"-o", outfile,
			infile,
		).Output()

		// 3. success
		if err == nil {
			fmt.Printf("ok\n")
			continue
		}

		// 4. debug error
		fmt.Printf("error\n")
		if len(stdout) > 0 {
			fmt.Printf("%s\n%s\n%s\n", clifmt.Color(31, "-=-"), stdout, clifmt.Color(31, "-=-"))
		}

	}

	return nil
}
