package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xdrm-io/clifmt"
)

// Arguments reads cli arguments from the terminal
type Arguments struct {
	ConfigPath    string
	GenFolderPath string
}

// Parse arguments from cli arguments
func (a *Arguments) Parse() error {
	flag.Usage = a.usage
	flag.Parse()

	if flag.NArg() != 2 {
		return fmt.Errorf("expected 2 arguments. Use -h for help")
	}

	a.ConfigPath = flag.Arg(0)
	stat, err := os.Stat(a.ConfigPath)
	if err != nil {
		return fmt.Errorf("cannot stat config file: %s", err)
	}
	if !stat.Mode().IsRegular() {
		return fmt.Errorf("config is not file")
	}
	if stat.Mode().Perm()&0400 == 0 {
		return fmt.Errorf("cannot read config file ; check your permissions")
	}

	a.GenFolderPath = flag.Arg(1)
	stat, err = os.Stat(a.GenFolderPath)
	if err == nil && !stat.IsDir() {
		return fmt.Errorf("generation folder already exists and is not a folder: %s", err)
	}
	if os.IsNotExist(err) {
		// create dir
		if err := os.Mkdir(a.GenFolderPath, 0777); err != nil {
			return fmt.Errorf("cannot create generation folder: %s", err)
		}
	}
	return nil
}

// usage notice when using -h or an invalid argument
func (Arguments) usage() {
	clifmt.Printf("**NAME**\n")
	clifmt.Printf("\tGenerates GO interfaces for your aicra REST API\n")
	clifmt.Printf("\n")
	clifmt.Printf("**SYNOPSIS**\n")
	clifmt.Printf("\t**aicragen** <config> <output_dir>\n")
	clifmt.Printf("\n")
	clifmt.Printf("**DESCRIPTION**\n")
	clifmt.Printf("\t**config**\n")
	clifmt.Printf("\t\tThe path to the api definition file (json) containing all endpoints.\n")
	clifmt.Printf("\n")
	clifmt.Printf("\t**output_dir**\n")
	clifmt.Printf("\t\tThe path of a folder for the GO files to be generated in.\n")
	clifmt.Printf("\n")
}
