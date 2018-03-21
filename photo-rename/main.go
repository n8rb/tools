package main

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const format = "20060102-150405-MST"

func main() {

	// Counters for summarizing actions taken
	var renamed, unchanged, errored int

	// If no arguments were passed, exit
	if len(os.Args) < 2 {
		log.Fatal("must specify a directory")
	}

	dir := os.Args[1]

	// List files in directory
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {

		// Full path to the source file
		path := filepath.Join(dir, file.Name())

		// Try to open
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("failed to open \"%v\"\n", file.Name())
			errored++
			continue
		}

		// Try to parse the EXIF data from the photo. Not all photos have this.
		// This will error if the info is missing, or if the file is not a photo.
		x, err := exif.Decode(f)
		f.Close()
		if err != nil {
			fmt.Printf("failed to decode file \"%v\": %v\n", file.Name(), err)
			errored++
			continue
		}

		dttm, err := x.DateTime()
		if err != nil {
			fmt.Printf("failed to extract date and time from \"%v\": %v\n", file.Name(), err)
			errored++
			continue
		}

		// Compile new filename based on extracted date
		newName := fmt.Sprintf("%v%v", dttm.Format(format), filepath.Ext(file.Name()))

		// If the file is already named correctly, do nothing.
		if file.Name() == newName {
			unchanged++
			continue
		}

		// Add directory to new filename
		newPath := filepath.Join(dir, newName)

		// Check if a file by the same name already exists.
		// Enhancement: if a file does already exist with the same name,
		// append a few digits from the digest of this file's contents.
		if _, err := os.Stat(newPath); err == nil {
			fmt.Printf("cannot rename %v to %v. file already exists with same name.\n", file.Name(), newName)
			errored++
			continue
		}

		// Rename the file
		err = os.Rename(path, newPath)
		if err != nil {
			fmt.Printf("failed to rename %v to %v: %v\n", file.Name(), newName, err)
			errored++
			continue
		}
		fmt.Printf("%v ---> %v\n", file.Name(), newName)
		renamed++
	}

	// Print a summary of actions taken
	fmt.Printf("%v files renamed, %v files unchanged, %v files errored\n", renamed, unchanged, errored)
}
