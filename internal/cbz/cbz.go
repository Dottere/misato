package cbz

import (
	"archive/zip"
	"io"
	"net/http"
	"strings"
)

type Cbz struct {
	handle *zip.ReadCloser
	// each index is a page number, each value is an index of an image
	// representing a page in the handle.File slice
	fileIndicesToPages []uint
}

func detectFileMimeType(f *zip.File, buf []byte) (string, error) {
	r, err := f.Open()
	if err != nil {
		return "", err
	}
	defer r.Close()

	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}

func OpenCbz(name string) (Cbz, error) {
	r, err := zip.OpenReader(name)
	if err != nil {
		return Cbz{}, err
	}

	cbz := Cbz{
		handle: r,
		// preallocate as many mappings as there are files
		fileIndicesToPages: make([]uint, len(r.File)),
	}

	const MAX_BYTES_TO_READ uint = 512
	buf := make([]byte, MAX_BYTES_TO_READ)
	nImages := 0

	for i, f := range r.File {
		mime, err := detectFileMimeType(f, buf)
		if err != nil {
			r.Close()
			return Cbz{}, err
		}
		if strings.HasPrefix(mime, "image/") {
			cbz.fileIndicesToPages[nImages] = uint(i)
			nImages = nImages + 1
		}
	}

	// shrink slice of mappings
	cbz.fileIndicesToPages = cbz.fileIndicesToPages[:nImages]
	return cbz, nil
}
