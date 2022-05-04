package main

// Dependencies
import (
	"errors"
	"fmt"
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
)

var (
	errDirectoryCreation = errors.New("FAILED TO CREATE DIRECTORY")
	errGetDirectory      = errors.New("FAILED TO GET THE CURRENT DIRECTORY")
	errCloneRepo         = errors.New("FAILED TO CLONE THE REPOSITORY")
)

func main() {

	// Get the current directory
	currentDir, err := getCurrentDir()

	// Check for errors
	if err != nil {
		log.Panic(err)
	}

	// Print the current directory
	fmt.Println("Current directory: ", currentDir)

	// Create the directory
	dirName, err := createDirectory("Root")

	if err != nil {
		log.Panic(err)
	}

	// Print the message
	fmt.Println("Directory created:", currentDir+"/"+dirName)
	dirFullPath := currentDir + "/" + dirName

	// Call the clone repo function
	msg, err := cloneRepo(dirFullPath)

	// Check for errors
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(msg)
}

// The function to get the current working directory
func getCurrentDir() (string, error) {
	// Get the current directory
	pwd, err := os.Getwd()
	// Check for errors
	if err != nil {
		return "Failed", errGetDirectory
	}

	return pwd, err
}

// The function to create a directory
func createDirectory(name string) (string, error) {

	// Create the directory
	err := os.Mkdir(name, os.ModePerm)

	// Check for errors
	if err != nil {
		return "Failed to Create Directory", errDirectoryCreation
	}
	return name, nil
}

func cloneRepo(path string) (string, error) {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:        "https://github.com/algo7/TripAdvisor-Review-Scraper.git",
		RemoteName: "origin",
		Progress:   os.Stdout,
	})

	if err != nil {
		return "Failed", errCloneRepo
	}

	return "Repo cloned", nil

}
