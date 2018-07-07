package main

import (
	"fmt"
	"git.xdrm.io/go/aicra/clifmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Builds controllers as plugins (.so)
// from the sources in the @in folder
// recursively and generate .so files
// into the @out folder with the same structure
func buildControllers(in string, out string) error {

	/* (1) Create build folder */
	clifmt.Align("    . create output folder")
	err := os.MkdirAll(out, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Printf("ok\n")

	/* (1) List recursively */
	sources := []string{}
	err = filepath.Walk(in, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, "i.go") {
			sources = append(sources, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	/* (2) Print files */
	for _, infile := range sources {

		// 1. process output file name
		rel, _ := filepath.Rel(in, infile)
		outfile := strings.Replace(rel, ".go", ".so", 1)
		outfile = fmt.Sprintf("%s/%s", out, outfile)

		clifmt.Align(fmt.Sprintf("    . compile %s", clifmt.Color(33, rel)))

		// 3. compile
		stdout, err := exec.Command("go",
			"build", "-buildmode=plugin",
			"-o", outfile,
			infile,
		).Output()

		// 4. success
		if err == nil {
			fmt.Printf("ok\n")
			continue
		}

		// 5. debug error
		fmt.Printf("error\n")
		if len(stdout) > 0 {
			fmt.Printf("%s\n%s\n%s\n", clifmt.Color(31, "-=-"), stdout, clifmt.Color(31, "-=-"))
		}

	}

	return nil
}
