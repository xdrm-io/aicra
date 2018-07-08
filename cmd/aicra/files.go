package main

import (
	"os"
	"path/filepath"
)

// Returns an absolute path from the @path variable if already absolute
// if @path is relative, it is processed relative to the @base directory
func getAbsPath(base string, path string) (string, error) {

	// already absolute
	if filepath.IsAbs(path) {
		return path, nil
	}

	// relative: join from @base dir
	return filepath.Abs(filepath.Join(base, path))
}

// Returns whether a directory exists for the path @path
func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err != nil || !stat.IsDir()
}
