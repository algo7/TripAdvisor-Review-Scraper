package utils

import (
	"archive/tar"
	"io"
	"os"
)

func WriteToFile(filename string, tarF io.ReadCloser) error {

	// Create the file
	out, err := os.Create(filename)
	defer out.Close()
	if err != nil {
		return err
	}

	// Untar the file
	// Note: This is not a generic untar function. It only works for a single file
	/**
		A tar file is a collection of binary data segments (usually sourced from files). Each segment starts with a header that contains metadata about the binary data, that follows it, and how to reconstruct it as a file.

	+---------------------------+
	| [name][mode][uid][guild]  |
	| ...                       |
	+---------------------------+
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	+---------------------------+
	| [name][mode][uid][guild]  |
	| ...                       |
	+---------------------------+
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	| XXXXXXXXXXXXXXXXXXXXXXXXX |
	+---------------------------+
		**/

	// Read the tar file
	tarReader := tar.NewReader(tarF)

	// Go to the next entry in the tar file
	_, err = tarReader.Next()

	if err != nil {
		return err
	}

	// Write the file to disk
	_, err = io.Copy(out, tarReader)
	if err != nil {
		return err
	}

	return err
}
