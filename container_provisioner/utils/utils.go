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

	// Untar the file
	// Note: This is not a generic untar function. It only works for a single file
	tarReader := tar.NewReader(tarF)
	io.Copy(out, tarReader)

	return err
}
