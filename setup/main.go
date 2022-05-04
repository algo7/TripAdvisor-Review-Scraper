package main

// Dependencies
import (
	"errors"
	"fmt"
	"os"

	git "github.com/go-git/go-git/v5"
)

var directoryCreationFailure = errors.New("Failed to create directory")
var getDirectoryFailure = errors.New("Failed to get the current directory")

func main() {

	// Get the current directory
	currentDir, err := getCurrentDir()

	if err != nil {
		panic(err)
	}

	// Call the clone repo function
	cloneRepo(currentDir)
}

// The function to get the current working directory
func getCurrentDir() (string, error) {
	// Get the current directory
	pwd, err := os.Getwd()
	// Check for errors
	if err != nil {
		return "Failed", getDirectoryFailure
	}
	// Print the current directory
	fmt.Println("Current directory:", pwd)

	return pwd, err
}

func cloneRepo(path string) {
	r, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:        "https://github.com/algo7/TripAdvisor-Review-Scraper.git",
		RemoteName: "origin",
		Progress:   os.Stdout,
	})

	if err != nil {
		fmt.Println(err)
	}

}
