package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	mapFileNames = make(map[string]bool)

	isRecursive        = false // flag for recursive directory scanning
	isCreateFiles      = true  // flag to enable/disable file creation
	isAccessTimeOnly   = false // flag to update last access time only
	isModifiedTimeOnly = false // flag to update last modified time only

	// defaults for modified and accessed time
	timeModified = time.Now().Local()
	timeAccessed = timeModified

	version = "0.4.0"
)

func main() {

	// get number of command line args and check if we have the minimum
	var numArgs = len(os.Args)
	if numArgs < 2 {
		printHelp()
		return
	}

	// go through the provided command line arguments and set the relevant
	// switches or build an array of unique file/directory names
	for i := 1; i < numArgs; i++ {
		filePattern := os.Args[i]

		// check command line arguments
		if strings.HasPrefix(filePattern, "-") {
			// check if help was requested
			if filePattern == "-h" || filePattern == "--help" {
				printHelp()
				return
			}

			// check if version info was requested
			if filePattern == "-v" || filePattern == "--version" {
				fmt.Printf("%s version %s\n", os.Args[0], version)
				return
			}

			// check if the recusrsive argument has been passed along
			if filePattern == "-R" || filePattern == "--recursive" {
				isRecursive = true
				continue
			}

			// check if file creation should be disabled
			if filePattern == "-c" || filePattern == "--no-create" {
				isCreateFiles = false
				continue
			}

			// check if only file modification time should be changed
			if filePattern == "-m" {
				isModifiedTimeOnly = true
				continue
			}

			// check if only file access time should be changed
			if filePattern == "-a" {
				isAccessTimeOnly = true
				continue
			}

			// get modified/accessed time from a reference file rather than
			// using current time
			if strings.HasPrefix(filePattern, "-r=") || strings.HasPrefix(filePattern, "--reference=") {
				// this looks weird, but so is PowerShell. this is a workaround
				// for how PowerShell passes arguments to programs.
				var refFilename = ""
				if filePattern == "-r=" || filePattern == "--reference=" {
					if i++; i < numArgs {
						refFilename = os.Args[i]
					}
				} else {
					refFilename = strings.Split(filePattern, "=")[1]
				}

				if refFilename == "" {
					log.Fatal("Reference file not provided")
				}

				timeMod, timeAcs, err := getFileTimes(refFilename)
				if os.IsNotExist(err) {
					log.Fatalf("Reference file %s not found", refFilename)
				}
				timeModified, timeAccessed = timeMod, timeAcs
				continue
			}

			// the below are mutually exclusive, so exit if both set
			if isAccessTimeOnly && isModifiedTimeOnly {
				log.Fatal("Access-time-only and modified-time-only flags are mutually exclusive")
			}

			log.Fatalf("Unknown argument %s\n", filePattern)
		}

		// expand the filepattern
		fileNames, err := filepath.Glob(filePattern)
		if err != nil {
			log.Println(err)
			continue
		}

		// if array is empty we'll assume it's a new filename that needs to be
		// created, so we push it to the array
		if len(fileNames) == 0 {
			addFileName(filePattern)
		} else {
			for _, fileName := range fileNames {
				addFileName(fileName)
			}
		}
	}

	// if the recursive flag is set then go through the array of names and
	// check if any are directories and add all the files inside them to
	// the array
	if isRecursive {
		// build a list of directories from the array of names
		var dirNames = []string{}
		for fileName := range mapFileNames {
			if isDir, _ := isDirectory(fileName); isDir {
				dirNames = append(dirNames, fileName)
			}
		}

		// now use Walk to grab all files and directories in the provided
		// folders, add them to the filenames array and ignore all errors
		for _, dirName := range dirNames {
			err := filepath.Walk(dirName, func(path string, info fs.FileInfo, err error) error {
				// ignore errors for now, and just pass them along
				if err != nil {
					log.Println(err)
				} else {
					addFileName(path)
				}
				return err
			})
			if err != nil {
				log.Println(err)
			}
		}
	}

	// finally go through the array of unique file paths and update/create them
	if len(mapFileNames) == 0 {
		log.Fatal("No files or directories provided")
	} else {
		for fileName := range mapFileNames {
			touch(fileName)
		}
	}
}

// print help text
func printHelp() {
	fmt.Printf("Usage: %s [-h|--help][-v|--version][-R|--recursive][-c|--no-create][-r=FILE|--reference=FILE][-a][-m] <paths> ...", os.Args[0])
}

// add string to array and make sure it's unique
func addFileName(strFileName string) {
	mapFileNames[strFileName] = true
}

// checks if a file exists. if it does it changes the timestamps, but if it
// doesn't exist it creates a file
func touch(fileName string) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		createFile(fileName)
	} else {
		changeFileTime(fileName)
	}
}

// simply creates a new file
func createFile(fileName string) {
	if isCreateFiles {
		file, err := os.Create(fileName)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
	}
}

// changes the access and modification timestamp of a file to the current time
func changeFileTime(fileName string) {
	targetAccessedTime, targetModifiedTime := timeAccessed, timeModified

	if isAccessTimeOnly {
		mad, _, err := getFileTimes(fileName)
		if err != nil {
			log.Println(err)
		} else {
			targetModifiedTime = mad
		}
	} else if isModifiedTimeOnly {
		_, tad, err := getFileTimes(fileName)
		if err != nil {
			log.Println(err)
		} else {
			targetAccessedTime = tad
		}
	}

	err := os.Chtimes(fileName, targetAccessedTime, targetModifiedTime)
	if err != nil {
		log.Println(err)
	}
}

// get last modified and last accessed times for a file
func getFileTimes(fileName string) (time.Time, time.Time, error) {
	fileInfo, err := os.Stat(fileName)
	// if error then return default time values and error value
	if err != nil {
		return timeModified, timeAccessed, err
	}
	return fileInfo.ModTime(), fileInfo.ModTime(), nil
}

// determine if a file represented by `path` is a directory or not
func isDirectory(path string) (bool, error) {
	if fileInfo, err := os.Stat(path); err != nil {
		return false, err
	} else {
		return fileInfo.IsDir(), nil
	}
}
