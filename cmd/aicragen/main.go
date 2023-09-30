package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/xdrm-io/aicra/codegen"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/clifmt"
)

func main() {
	var args Arguments
	if err := args.Parse(); err != nil {
		clifmt.Fprintf(os.Stderr, "${invalid argument}(red) | %s\n", err)
		os.Exit(1)
	}

	configFile, err := os.ReadFile(args.ConfigPath)
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${config}(red) | %s\n", err)
		os.Exit(1)
	}

	// parse config
	var cnf config.API
	if err := json.Unmarshal(configFile, &cnf); err != nil {
		clifmt.Fprintf(os.Stderr, "${config metadata}(red) | %s\n", err)
		os.Exit(1)
	}

	var gen = codegen.Generator{Config: cnf}

	f, err := os.OpenFile(filepath.Join(args.GenFolderPath, "validators.go"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${generate validators}(red) | %s\n", err)
		os.Exit(1)
	}
	err = gen.WriteValidators(f)
	f.Close()
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${generate validators}(red) | %s\n", err)
		os.Exit(1)
	}

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
	f, err = os.OpenFile(filepath.Join(args.GenFolderPath, "mappers.go"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${generate mappers}(red) | %s\n", err)
		os.Exit(1)
	}
	err = gen.WriteMappers(f)
	f.Close()
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${generate mappers}(red) | %s\n", err)
		os.Exit(1)
	}

	log.Printf("--GENERATED--")
}
