package main

import (
	"flag"
	"fmt"
	"git.xdrm.io/go/aicra/clifmt"
	"os"
	"path/filepath"
)

func main() {

	/* (1) Flags
	---------------------------------------------------------*/
	/* (1) controller path */
	ctlPathFlag := flag.String("c", "root", "Path to controllers' directory")

	/* (2) types path */
	typPathFlag := flag.String("t", "custom-types", "Path to custom types' directory")

	flag.Parse()

	/* (3) Get last arg: project path */
	if len(flag.Args()) < 1 {
		fmt.Printf("%s\n\n", clifmt.Warn("missing argument"))
		fmt.Printf("You must provide the project folder as the last argument\n")
		return
	}
	var projectPathFlag = flag.Arg(0)
	compileTypes := true
	compileControllers := true

	/* (2) Get absolute paths
	---------------------------------------------------------*/
	/* (1) Get absolute project path */
	projectPath, err := filepath.Abs(projectPathFlag)
	if err != nil {
		fmt.Printf("invalid argument: project path\n")
		return
	}
	/* (2) Get absolute controllers' path */
	if !filepath.IsAbs(*ctlPathFlag) {
		*ctlPathFlag = filepath.Join(projectPath, *ctlPathFlag)
	}
	cPath, err := filepath.Abs(*ctlPathFlag)
	if err != nil {
		fmt.Printf("invalid argument: controllers' path\n")
		return
	}

	/* (3) Get absolute types' path */
	if !filepath.IsAbs(*typPathFlag) {
		*typPathFlag = filepath.Join(projectPath, *typPathFlag)
	}
	tPath, err := filepath.Abs(*typPathFlag)
	if err != nil {
		fmt.Printf("invalid argument: types' path\n")
		return
	}

	/* (3) Check path are existing dirs
	---------------------------------------------------------*/
	clifmt.Title("Check files")

	/* (1) Project path */
	clifmt.Align("   . project root")
	if stat, err := os.Stat(projectPath); err != nil || !stat.IsDir() {
		fmt.Printf("invalid\n\n")
		fmt.Printf("%s  invalid project folder - %s\n\n", clifmt.Warn(), clifmt.Color(36, projectPath))
		fmt.Printf("You must specify an existing directory path\n")
		return
	} else {
		fmt.Printf("ok\n")
	}

	/* (2) Controllers path */
	clifmt.Align("   . controllers")
	if stat, err := os.Stat(cPath); err != nil || !stat.IsDir() {
		compileControllers = false
		fmt.Printf("missing\n")
	} else {
		fmt.Printf("ok\n")
	}

	/* (3) Types path */
	clifmt.Align("   . custom types")
	if stat, err := os.Stat(tPath); err != nil || !stat.IsDir() {
		fmt.Printf("missing\n")
		compileTypes = false

	} else {
		fmt.Printf("ok\n")
	}

	if !compileControllers && !compileTypes {
		fmt.Printf("\n%s\n", clifmt.Info("Nothing to compile"))
		return
	}

}
