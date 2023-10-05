package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xdrm-io/aicra/codegen"
	"github.com/xdrm-io/aicra/internal/config"
	"github.com/xdrm-io/clifmt"
)

func main() {
	start := time.Now()

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

	var (
		vPath = filepath.Join(args.GenFolderPath, "validators.go")
		ePath = filepath.Join(args.GenFolderPath, "endpoints.go")
		mPath = filepath.Join(args.GenFolderPath, "mappers.go")
	)

	// relative path of config from the generated code
	configRelPath, err := filepath.Rel(args.GenFolderPath, args.ConfigPath)
	if err != nil {
		clifmt.Fprintf(os.Stderr, "${config path}(red) | %s\n", err)
		os.Exit(1)
	}

	var gen = codegen.Generator{Config: cnf, ConfigRelPath: configRelPath}

	f, err := os.OpenFile(vPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
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

	f, err = os.OpenFile(ePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
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
	f, err = os.OpenFile(mPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
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

	clifmt.Printf("> ${successfully generated}(#19c681)\n")
	clifmt.Printf("- ${%s}(gray)\n", strings.ReplaceAll(vPath, `_`, `\_`))
	clifmt.Printf("- ${%s}(gray)\n", strings.ReplaceAll(ePath, `_`, `\_`))
	clifmt.Printf("- ${%s}(gray)\n", strings.ReplaceAll(mPath, `_`, `\_`))
	clifmt.Printf("\ntook ${%s}(gray)\n", time.Since(start))
}
