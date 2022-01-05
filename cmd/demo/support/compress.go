package support

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Compress(writer http.ResponseWriter, path string) {
	gzipWriter := gzip.NewWriter(writer)
	tarWriter := tar.NewWriter(gzipWriter)
	err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(path, file)
		if err != nil {
			return err
		}
		header.Name = rel

		if err := tarWriter.WriteHeader(header); err != nil {
			fmt.Println(err)
			return err
		}
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	if err := tarWriter.Close(); err != nil {
		fmt.Println("unable to tar.")
	}
	if err := gzipWriter.Close(); err != nil {
		fmt.Println("unable to gzip.")
	}
}
