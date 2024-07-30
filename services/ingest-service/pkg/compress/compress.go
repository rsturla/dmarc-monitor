package compress

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

func Decompress(data []byte, mimeType string) ([]byte, error) {
	switch mimeType {
	case "application/gzip":
		return decompressGzip(data)
	case "application/zip":
		return decompressZip(data)
	default:
		return nil, fmt.Errorf("unsupported MIME type: %s", mimeType)
	}
}

func decompressGzip(data []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()

	return io.ReadAll(gzipReader)
}

func decompressZip(data []byte) ([]byte, error) {
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
