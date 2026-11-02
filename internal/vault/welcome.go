package vault

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"godraw/internal/doc"
)

// Welcome writes the drawing a new folder starts with: a page that says
// what the folder is, and how to begin. A folder that was found, rather
// than made now, keeps what it has.
func (v *Vault) Welcome() error {
	if !v.fresh {
		return nil
	}
	list, err := v.Drawings()
	if err != nil {
		return err
	}
	if len(list) > 0 {
		return nil
	}
	var buf bytes.Buffer
	if err := doc.Write(&buf, welcome()); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(v.dir, "Welcome"+Ext), buf.Bytes(), 0o644)
}

// welcome is the drawing: a title, a line about the folder, and three
// notes that say what to do next.
func welcome() *doc.Drawing {
	d := doc.New()
	page := d.Pages[0].ID
	last := ""
	add := func(s *doc.Shape) {
		s.Parent = page
		s.Index = doc.IndexBetween(last, "")
		last = s.Index
		d.Shapes[s.ID] = s
	}

	title := doc.NewShape("shape:welcome-title", doc.TypeText)
	title.X, title.Y = 0, 0
	title.Text = "Welcome to godraw"
	title.Size, title.Scale, title.Color = "xl", 2.4, "black"
	add(title)

	sub := doc.NewShape("shape:welcome-sub", doc.TypeText)
	sub.X, sub.Y = 6, 112
	sub.Text = "A whiteboard that lives in one folder. This one."
	sub.Size, sub.Color = "l", "grey"
	add(sub)

	notes := []struct {
		x     float64
		color string
		text  string
	}{
		{0, "light-violet", "Draw anything\nD draws, R makes shapes, T types."},
		{272, "light-blue", "It saves itself\nEvery change is written to this folder."},
		{544, "yellow", "Yours to keep\nEach drawing is a .tldr file, as tldraw's."},
	}
	for i, n := range notes {
		s := doc.NewShape(doc.ID(fmt.Sprintf("shape:welcome-note-%d", i)), doc.TypeNote)
		s.X, s.Y = float64(n.x), 208
		s.W, s.H = 248, 208
		s.Color, s.Text, s.Size = n.color, n.text, "m"
		add(s)
	}
	return d
}
