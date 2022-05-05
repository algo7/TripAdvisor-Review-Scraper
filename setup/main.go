package main

// Dependencies
import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	git "github.com/go-git/go-git/v5"
)

// Custom errors
var (
	errDirectoryCreation        = errors.New("FAILED TO CREATE DIRECTORIES")
	errGetDirectory             = errors.New("FAILED TO GET THE CURRENT DIRECTORY")
	errPurgeDirectory           = errors.New("FAILED TO PURGE THE TMP DIRECTORY")
	errCopyFile                 = errors.New("FAILED TO COPY DOCKER-COMPSE-PROD.YML")
	errCloneRepo                = errors.New("FAILED TO CLONE THE REPOSITORY")
	errDockerCheck              = errors.New("DOCKER IS NOT INSTALLED")
	errSetupCheck               = errors.New("SETUP CHECK FAILED")
	errDockerComposeRun         = errors.New("FAILED TO RUN DOCKER-COMPOSE")
	errReviewsNotEmpty          = errors.New("REVIEWS DIRECTORY IS NOT EMPTY")
	errMissingSourceFiles       = errors.New("MISSING SOURCE FILES")
	errInputScrapMode           = errors.New("INVALID SCRAP MODE")
	errInputConcurrency         = errors.New("INVALID CONCURRENCY VALUE")
	errDockerComposeYmlNotFound = errors.New("DOCKER-COMPSE-PROD.YML NOT FOUND")
	errValueReplace             = errors.New("FAILED TO REPLACE VALUE")
)

// The main function
func main() {

	// Check if docker is installed
	msg, err := checkDocker()
	errorHandler(err)
	fmt.Println("0. " + msg)

	// Get the current directory
	currentDir, err := getCurrentDir()

	// Check for errors
	errorHandler(err)

	err = userInputs(currentDir)
	errorHandler(err)

	// Print the current directory
	fmt.Println("1. Current directory: ", currentDir)

	// Run the setup check
	isCompleted, err := setupCheck(currentDir)

	// Check for errors
	errorHandler(err)

	// If the setup is completed already, run the docker container
	if isCompleted {
		err := dockerComposeRun(filepath.Join(currentDir, "Project_Files"))
		errorHandler(err)
		return
	}

	// Create a temporary directory to hold the repository
	tmpDirName, err := createDirectory("tmp")

	// Check for errors
	errorHandler(err)

	tmpDirFullPath := filepath.Join(currentDir, tmpDirName)

	// Print the message
	fmt.Println("2. Tmp Directory created:", tmpDirFullPath)

	// Create a temporary directory to hold the repository
	projectDirName, err := createDirectory("Project_Files")

	// Check for errors
	errorHandler(err)

	projectDirFullPath := filepath.Join(currentDir, projectDirName)

	// Print the message
	fmt.Println("3. Project Directory created:", projectDirFullPath)

	// Call the clone repo function
	msg, err = cloneRepo(tmpDirFullPath)

	// Check for errors
	errorHandler(err)
	fmt.Println("4. " + msg)

	// Copy docker-compose-prod.yml to the Project_Files directory
	msg, err = copy(
		filepath.Join(tmpDirFullPath, "docker-compose-prod.yml"),
		filepath.Join(projectDirFullPath, "docker-compose-prod.yml"))

	// Check for errors
	errorHandler(err)
	fmt.Println("5. " + msg)

	// Purge the temporary directory
	msg, err = purgeDir(tmpDirFullPath)
	// Check for errors
	errorHandler(err)
	fmt.Println("6. " + msg)

	// Create the source directory
	sourceDirFullPath := filepath.Join(projectDirFullPath, "source")
	_, err = createDirectory(sourceDirFullPath)

	// Check for errors
	errorHandler(err)

	// Print the message
	fmt.Println("7. Source Directory created:", sourceDirFullPath)

	// Create the reviews directory
	reviewsDirFullPath := filepath.Join(projectDirFullPath, "reviews")
	_, err = createDirectory(reviewsDirFullPath)

	// Check for errors
	errorHandler(err)

	// Print the message
	fmt.Println("7. Reviews Directory created:", reviewsDirFullPath)

}

func userInputs(path string) error {

	// Get scrap mode
	fmt.Println("Enter the scrap mode:")
	var mode string
	_, err := fmt.Scanf("%s", &mode)

	// Input validation
	if err != nil || (mode != "HOTEL" && mode != "RESTO") {
		return errInputScrapMode
	}

	// Get concurrency value
	fmt.Println("Enter the concurrency value:")
	var i int
	_, err = fmt.Scanf("%d", &i)

	// Input validation
	if err != nil {
		return errInputConcurrency
	}

	// Read the docker-compose-prod.yml file
	dockerComposeFilePath := filepath.Join(path, "Project_Files/docker-compose-prod.yml")
	content, err := ioutil.ReadFile(dockerComposeFilePath)

	if err != nil {
		return errDockerComposeYmlNotFound
	}

	// Regex to match the scrap mode
	scrapModeRegex := regexp.MustCompile("SCRAPE_MODE:(.*)")
	// Regex to match the concurrency value
	concurrencyRegex := regexp.MustCompile("CONCURRENCY:(.*)")

	// Replace the scrap mode with the input
	scrapModeChanged := scrapModeRegex.ReplaceAllString(string(content), "SCRAPE_MODE: "+mode)
	// Replace the concurrency value with the input
	concurrencyChanged := concurrencyRegex.ReplaceAllString(scrapModeChanged, "CONCURRENCY: "+strconv.Itoa(i))

	f, err := os.OpenFile(dockerComposeFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return errValueReplace
	}

	/*defer can be effectively used to release critical system resources – such as closing an open file – to ensure that our code does not leak file descriptors.*/
	defer f.Close()

	// Write the new content to the file
	_, err = f.WriteString(concurrencyChanged)

	if err != nil {
		return errValueReplace
	}

	return nil
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

// Clone the repository
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

// Copy the file from source to destination
func copy(sourceFile string, destFile string) (string, error) {

	// Read the source file
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return "Ops", errCopyFile
	}

	// Write tp the destination file
	err = ioutil.WriteFile(destFile, input, os.ModePerm)
	if err != nil {
		return "Ops", errCopyFile
	}

	return "docker-compose-prod.yml copied successfully", nil
}

// Remove all the directories and files given the path
func purgeDir(path string) (string, error) {

	err := os.RemoveAll(path)
	if err != nil {
		return "Ops", errPurgeDirectory
	}
	return "Tmp directory purged", nil
}

// Check if docker is installed
func checkDocker() (string, error) {
	cmd := exec.Command("docker", "-v")

	err := cmd.Run()

	if err != nil {
		return "Ops", errDockerCheck
	}

	return "Docker is installed", nil
}

/* Check if the setup process has been completed already
* If it has, spin up the docker container
 */
func setupCheck(path string) (bool, error) {

	// Get all the directories and files in the current directory
	files, err := ioutil.ReadDir(path)

	if err != nil {
		return false, errSetupCheck
	}

	var projectDirExists = false

	// Check if the Project_Files directory exists
	for _, file := range files {
		if file.Name() == "Project_Files" && file.IsDir() == true {
			projectDirExists = true
		}
	}

	// If the project directory does not exist, return false
	if !projectDirExists {
		return false, nil
	}

	// Get all the directories and files in the project directory
	projectFileDir := filepath.Join(path, "Project_Files")
	files, err = ioutil.ReadDir(projectFileDir)

	if err != nil {
		return false, errSetupCheck
	}

	// Check if the source and reviews directory exists
	var sourceDirExists = false
	var reviewsDirExists = false

	for _, file := range files {
		if file.Name() == "source" && file.IsDir() == true {
			sourceDirExists = true
		}
		if file.Name() == "reviews" && file.IsDir() == true {
			reviewsDirExists = true
		}
	}

	// If the source and reviews directories do not exist, return false
	if !sourceDirExists || !reviewsDirExists {
		return false, nil
	}

	// Get all the directories and files in the project directory
	sourceFiles, err := ioutil.ReadDir(filepath.Join(projectFileDir, "source"))
	if err != nil {
		return false, errSetupCheck
	}
	reviewFiles, err := ioutil.ReadDir(filepath.Join(projectFileDir, "reviews"))
	if err != nil {
		return false, errSetupCheck
	}

	// Check if the restos.csv / hotel.csv files exist in the source directory
	var sourceCSVExists = false

	// Check if the source files exist
	for _, sourceFile := range sourceFiles {
		if (sourceFile.Name() == "restos.csv" || sourceFile.Name() == "hotels.csv") && !sourceFile.IsDir() {
			sourceCSVExists = true
		}
	}

	if !sourceCSVExists {
		return false, errMissingSourceFiles
	}

	// Check if the reviews folder is empty
	if len(reviewFiles) != 0 {
		return false, errReviewsNotEmpty
	}

	return true, nil
}

// Spin up the docker container
func dockerComposeRun(path string) error {

	// Prompt and wait for user input
	fmt.Println("Press Any Key to Run The Scraper")
	fmt.Scanln()

	// The path to the docker-compose file
	dockerComposePath := filepath.Join(path, "docker-compose-prod.yml")

	// Run the docker container
	cmd := exec.Command("docker-compose", "-f", dockerComposePath, "up")

	// Create a pipe that connects to the stdout of the command
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return errDockerComposeRun
	}

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan struct{})

	// Creates a scanner that read the stdout/stderr line-by-line
	scanner := bufio.NewScanner(stdout)

	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
		}

		// We're all done, unblock the channel
		done <- struct{}{}

	}()

	// Start the command but do not wait until it is finished
	err = cmd.Start()

	if err != nil {
		return errDockerComposeRun
	}

	err = cmd.Wait()

	if err != nil {
		return errDockerComposeRun
	}

	return nil
}

// Custom error handler
func errorHandler(err error) {
	if err != nil {
		fmt.Println(err)
		fmt.Println("Press Any Key to Exit")
		fmt.Scanln()
		os.Exit(0)
	}
}
