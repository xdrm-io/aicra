package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/xdrm-io/clifmt"
)

func main() {
	var args Arguments
	if err := args.Parse(); err != nil {
		clifmt.Fprintf(os.Stderr, "${invalid argument}(red) | %s\n", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(args.ConfigPath, os.O_RDONLY, 0400)
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${config}(red) | %s\n", err)
		os.Exit(1)
	}

	// parse config
	var cnf Config
	err = cnf.Decode(f)
	f.Close()
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${config metadata}(red) | %s\n", err)
		os.Exit(1)
	}

	var gen = Generator{cnf}

	f, err = os.OpenFile(filepath.Join(args.GenFolderPath, "endpoints.go"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${generate endpoints}(red) | %s\n", err)
		os.Exit(1)
	}
	err = gen.WriteEndpoints(f)
	f.Close()
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${generate endpoints}(red) | %s\n", err)
		os.Exit(1)
	}

	log.Printf("--GENERATED--")
}
