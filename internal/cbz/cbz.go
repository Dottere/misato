package cbz

import (
	"archive/zip"
	"net/url"
	"path/filepath"
	"strings"
)

type Cbz struct {
	Handle *zip.ReadCloser
	// each index is a page number, each value is an index of an image
	// representing a page in the handle.File slice
	FileIndicesToPages []uint
	UrlPath            string
	Title              string
}

func pathToUrl(path string) string {
	basename := filepath.Base(path)
	ext := filepath.Ext(basename)
	return url.PathEscape(strings.TrimSuffix(basename, ext))
}

func OpenCbz(name string) (Cbz, error) {
	r, err := zip.OpenReader(name)
	if err != nil {
		return Cbz{}, err
	}

	cbz := Cbz{
		Handle: r,
		// preallocate as many mappings as there are files
		FileIndicesToPages: make([]uint, len(r.File)),
		UrlPath:            pathToUrl(name),
		Title:              name,
	}

	nImages := 0

	for i, f := range r.File {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
			cbz.FileIndicesToPages[nImages] = uint(i)
			nImages++
		}
	}

	// shrink slice of mappings
	cbz.FileIndicesToPages = cbz.FileIndicesToPages[:nImages]
	return cbz, nil
}
