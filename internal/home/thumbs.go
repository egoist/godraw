package home

import (
	"os"
	"time"

	"github.com/egoist/mygo/ui"

	"godraw/internal/app"
	"godraw/internal/doc"
	"godraw/internal/vault"
)

// thumb is what the window knows of a drawing without opening it: a
// picture of its page, and what is on it.
type thumb struct {
	bmp    *ui.Bitmap
	mod    time.Time
	size   int64
	dark   bool
	pages  int
	shapes int
	read   bool // the file could be read
}

// The size of a picture, in DIPs, and how many device pixels a DIP is.
const (
	thumbW, thumbH = 400, 225
	thumbScale     = 2
	// thumbsPerFrame bounds what drawing pictures costs a frame: a folder
	// of heavy drawings fills in over a few of them, rather than stopping
	// the window as it opens.
	thumbsPerFrame = 2
)

// thumbOf returns the picture of a drawing, drawing it if the file is new
// or has changed since it was.
func (h *Home) thumbOf(d vault.Drawing, dark bool) *thumb {
	if t := h.thumbs[d.Path]; t != nil && t.mod.Equal(d.Mod) && t.size == d.Size && t.dark == dark {
		return t
	}
	if h.drawn >= thumbsPerFrame {
		h.waiting = true
		return h.thumbs[d.Path]
	}
	h.drawn++
	t := renderThumb(d, dark)
	h.thumbs[d.Path] = t
	return t
}

// renderThumb reads a drawing and draws it.
func renderThumb(d vault.Drawing, dark bool) *thumb {
	t := &thumb{mod: d.Mod, size: d.Size, dark: dark}
	f, err := os.Open(d.Path)
	if err != nil {
		return t
	}
	defer f.Close()
	dr, err := doc.Read(f)
	if err != nil {
		return t
	}
	t.read = true
	t.pages, t.shapes = len(dr.Pages), len(dr.Shapes)
	t.bmp = app.Preview(dr, thumbW, thumbH, thumbScale, dark)
	return t
}
