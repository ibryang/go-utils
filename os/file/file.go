package file

import (
	"path/filepath"
	"strings"
)

func Name(path string) string {
	base := filepath.Base(path)
	if i := strings.LastIndexByte(base, '.'); i != -1 {
		return base[:i]
	}
	return base
}
func Basename(path string) string {
	return filepath.Base(path)
}

func Ext(path string) string {
	ext := filepath.Ext(path)
	if p := strings.IndexByte(ext, '?'); p != -1 {
		ext = ext[0:p]
	}
	return ext
}

func ExtName(path string) string {
	return Ext(path)[1:]
}
