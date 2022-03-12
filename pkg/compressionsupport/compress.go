package compressionsupport

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func TarFromPath(path string) ([]byte, error) {
	var buffer bytes.Buffer
	tw := tar.NewWriter(&buffer)
	defer func(tw *tar.Writer) {
		_ = tw.Close()
	}(tw)

	err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		header, headerErr := tar.FileInfoHeader(fi, file)
		if headerErr != nil {
			return headerErr
		}

		rel, pathErr := filepath.Rel(path, file)
		if pathErr != nil {
			return pathErr
		}
		header.Name = rel

		headerErr = tw.WriteHeader(header)
		if headerErr != nil {
			return headerErr
		}

		if !fi.IsDir() {
			data, openErr := os.Open(file)
			if openErr != nil {
				return openErr
			}
			_, copyErr := io.Copy(tw, data)
			if copyErr != nil {
				return copyErr
			}
		}
		return nil
	})
	return buffer.Bytes(), err
}

func UnTarToPath(r io.Reader, path string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.Contains(hdr.Name,"..") {
			return errors.New("zip slip fix")
		}
		join := filepath.Join(path, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			if _, statErr := os.Stat(join); statErr != nil {
				if mkdirErr := os.MkdirAll(join, 0744); mkdirErr != nil {
					return mkdirErr
				}
			}
		}
		if hdr.Typeflag == tar.TypeReg {
			f, openErr := os.OpenFile(join, os.O_CREATE|os.O_RDWR, 0644)
			if openErr != nil {
				return openErr
			}
			if _, copyErr := io.Copy(f, tr); copyErr != nil {
				return copyErr
			}
		}
	}
	return nil
}

func Gzip(w io.Writer, by []byte) error {
	zw := gzip.NewWriter(w)
	defer func(zw *gzip.Writer) {
		_ = zw.Close()
	}(zw)
	_, err := zw.Write(by)
	return err
}

func UnGzip(r io.Reader) ([]byte, error) {
	var uncompressed bytes.Buffer
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer func(zr *gzip.Reader) {
		_ = zr.Close()
	}(zr)
	_, copyErr := io.Copy(&uncompressed, zr)
	if copyErr != nil {
		return nil, copyErr
	}
	return uncompressed.Bytes(), nil
}
