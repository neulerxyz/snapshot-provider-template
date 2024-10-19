// internal/snapshot/compress.go

package snapshot

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"

	"github.com/pierrec/lz4/v4"
)

func CompressDirectory(sourceDir, destFile string) error {
	tarfile, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	// Create LZ4 writer
	lz4Writer := lz4.NewWriter(tarfile)
	defer lz4Writer.Close()

	// Create a new tar writer
	tarWriter := tar.NewWriter(lz4Writer)
	defer tarWriter.Close()

	// Walk through the source directory
	err = filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// Update the header name to maintain directory structure
		header.Name, _ = filepath.Rel(sourceDir, file)

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If not a regular file, don't need to copy data
		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		// Copy file data to tar writer
		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}

		return nil
	})

	return err
}
