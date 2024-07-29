package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

func uncompress(data []byte, mime string) ([]byte, error) {
	switch mime {
	case "application/gzip":
		return uncompressGzip(data)
	case "application/zip":
		return uncompressZip(data)
	default:
		return nil, fmt.Errorf("unsupported MIME type: %s", mime)
	}
}

func uncompressGzip(data []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()

	return io.ReadAll(gzipReader)
}

func uncompressZip(data []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("error creating zip reader: %w", err)
	}

	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() {
			file, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("error opening file: %w", err)
			}
			defer file.Close()

			return io.ReadAll(file)
		}
	}

	return nil, fmt.Errorf("no files found in zip archive")
}
