package cbz

import (
	"archive/zip"
	"net/http"
	"io"
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


	n, err := io.ReadAll(io.ReadFull(r, buf))
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}

	return http.detectContentType(buf[:n])
}

func OpenCbz(name string) (Cbz, error) {
	r, err := zip.OpenReader(name)
	if err != nil {
		return {}, err
	}

	cbz := Cbz {
		handle: r,
		// preallocate as many mappings as there are files
		fileIndicesToPages: make([]uint, len(r.File))
	}

	const MAX_BYTES_TO_READ uint = 512
	buf := make([]byte, MAX_BYTES_TO_READ)
	nImages := 0

	for i, f := range r.File {
		mime, err := detectFileMimeType(f, buf)
		if err != nil {
			r.Close()
			return {}, err
		}
		if strings.hasPrefix(mime, "image/") {
			cbz.fileIndicesToPages[nImages] = i
			nImages = nImages + 1
		}
	}

	// shrink slice of mappings
	cbz.fileIndicesToPages = append([]uint, cbz.fileIndicesToPages[:nImages])
	return cbz
}
