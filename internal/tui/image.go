package tui

import (
	"path/filepath"
	"strings"

	termimg "github.com/blacktop/go-termimg"
)

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".gif": true, ".bmp": true, ".webp": true,
	".tiff": true, ".tif": true,
}

func isImageFile(name string) bool {
	return imageExts[strings.ToLower(filepath.Ext(name))]
}

func renderImagePreview(path string, w, h int) string {
	if w <= 4 || h <= 2 {
		return ""
	}
	img, err := termimg.Open(path)
	if err != nil {
		return ""
	}
	rendered, err := img.
		Width(w - 4).
		Height(h - 2).
		Scale(termimg.ScaleFit).
		Protocol(termimg.Halfblocks).
		Render()
	if err != nil {
		return ""
	}
	return rendered
}
