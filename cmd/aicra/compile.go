package main

import (
	"fmt"
	"git.xdrm.io/go/aicra/internal/clifmt"
	"os"
	"os/exec"
	"path/filepath"
)

// compile compiles the 'source' file into the 'build' path
func compile(source, build string) {

	// 2. Create folder
	clifmt.Align("    + create output folder")
	err := os.MkdirAll(filepath.Dir(build), os.ModePerm)
	if err != nil {
		fmt.Printf("fail\n")
		return
	}
	fmt.Printf("ok\n")

	// 3. Compile
	clifmt.Align("    + compile")
	stdout, err := exec.Command("go",
		"build", "-ldflags", "-s -w", "-buildmode=plugin",
		"-o", build,
		source,
	).Output()

	// 4. success
	if err == nil {
		fmt.Printf("ok\n")
		return
	}

	// 5. debug error
	fmt.Printf("error\n")
	if len(stdout) > 0 {
		fmt.Printf("%s\n%s\n%s\n", clifmt.Color(31, "-=-"), stdout, clifmt.Color(31, "-=-"))
	}

}
