package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var dir string
var data map[string][]byte
var re *regexp.Regexp

func main() {

	var err error

	// The regex for save variable names
	re = regexp.MustCompile(`[^a-zA-Z0-9]`)

	// To store file data temporarily
	data = make(map[string][]byte)

	// Parse params
	if len(os.Args)-1 != 4 {
		fmt.Println("Usage: go run file2source PACKAGE PREFIX DIRECTORY OUTFILE.")
		return
	}
	pkgName := os.Args[1]
	prefix := os.Args[2]
	directory := os.Args[3]
	outFile := os.Args[4]

	// Get the absolute directory so that we know the folder name
	// I don't think we need this anymore.
	dir, err = filepath.Abs(directory)
	if err != nil {
		fmt.Printf("Couldn't determine directory path from given DIRECTORY: %v\n", err)
		os.Exit(1)
	}

	// Ensure the directory exists
	if _, err := os.Stat(dir); err != nil {
		fmt.Printf("Failed to obtain info about this directory: %v\n", err)
		os.Exit(1)
	}

	// Process each file
	filepath.Walk(dir, handleFile)

	// Generate the output file

	outStr := fmt.Sprintf("package %v\n\n", pkgName)
	outStr += fmt.Sprintf("// Created with file2source\n\n")
	outStr += fmt.Sprintf("// Contains the following file data:\n")
	outStr += fmt.Sprintf("// Filename:Bytes:Constant\n")

	for filename, dat := range data {
		outStr += fmt.Sprintf("// %v:%v:%v\n",
			filename, len(dat), varFromFilename(filename, prefix))
	}
	outStr += fmt.Sprintf("\n")

	for filename, dat := range data {
		base64Dat := base64.StdEncoding.EncodeToString(dat)
		outStr += fmt.Sprintf("const %v string = \"%v\"\n",
			varFromFilename(filename, prefix), base64Dat)
	}

	err = ioutil.WriteFile(outFile, []byte(outStr), 0644)
	if err != nil {
		fmt.Printf("Couldn't write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created file %v.\n", outFile)

}

// Handles the files sent from filepath.Walk.
func handleFile(path string, info os.FileInfo, err error) error {

	// Only process if this is not a directory, and this is a
	// child (not descendant) of the user-provided directory.
	if !info.IsDir() && filepath.Dir(path) == dir {

		filename := filepath.Base(path)

		// Read the file
		fmt.Printf("Loading %v...\n", filename)
		dat, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed to read file %v: %v", filename, err)
			os.Exit(1)
		}

		// Store the file's bytes
		data[filename] = dat
	}

	return nil
}

// Returns a variable name by removing non-alphanumeric characters
// from the file name and adding a prefix.
func varFromFilename (filename, prefix string) string {
	return prefix + re.ReplaceAllString(filename, "")	
}
