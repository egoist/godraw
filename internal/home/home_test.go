package home

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"godraw/internal/doc"
	"godraw/internal/vault"
)

// testVault makes a folder holding one drawing per name, each with a word
// on it, so that it has a picture to show.
func testVault(t *testing.T, names ...string) *vault.Vault {
	t.Helper()
	v, err := vault.New(filepath.Join(t.TempDir(), "drawings"))
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range names {
		d := doc.New()
		s := doc.NewShape(doc.ID("shape:text"), doc.TypeText)
		s.Parent = d.Pages[0].ID
		s.Index = doc.FirstIndex
		s.X, s.Y = 0, 0
		s.Text, s.Size = name, "xl"
		d.Shapes[s.ID] = s

		f, err := os.Create(filepath.Join(v.Dir(), name+".tldr"))
		if err != nil {
			t.Fatal(err)
		}
		if err := doc.Write(f, d); err != nil {
			t.Fatal(err)
		}
		f.Close()
		when := time.Now().Add(-time.Duration(i) * time.Hour)
		os.Chtimes(filepath.Join(v.Dir(), name+".tldr"), when, when)
	}
	return v
}

func TestViewShowsTheFolder(t *testing.T) {
	h := New(testVault(t, "Ideas", "Wireframes"))
	tr := ui.NewTester(h.View, 1000, 700)

	for _, want := range []string{"Drawings", "Ideas", "Wireframes", "New drawing", "RECENT"} {
		if !tr.HasText(want) {
			t.Errorf("the window doesn't show %q: %v", want, tr.Texts())
		}
	}
}

func TestViewOfAnEmptyFolder(t *testing.T) {
	h := New(testVault(t))
	tr := ui.NewTester(h.View, 1000, 700)
	if !tr.HasText("Your folder is empty") {
		t.Errorf("the empty folder doesn't say so: %v", tr.Texts())
	}
}

func TestClickingACardOpensIt(t *testing.T) {
	h := New(testVault(t, "Ideas", "Wireframes"))
	var opened string
	h.Open = func(path string) { opened = path }

	tr := ui.NewTester(h.View, 1000, 700)
	if err := tr.Click("Ideas"); err != nil {
		t.Fatal(err)
	}
	tr.Frame()
	if want := filepath.Join(h.v.Dir(), "Ideas.tldr"); opened != want {
		t.Fatalf("opened %q, want %q", opened, want)
	}
}

func TestClickingNewStartsADrawing(t *testing.T) {
	h := New(testVault(t, "Ideas"))
	made := 0
	h.New = func() { made++ }

	tr := ui.NewTester(h.View, 1000, 700)
	if err := tr.Click("New drawing"); err != nil {
		t.Fatal(err)
	}
	tr.Frame()
	if made != 1 {
		t.Fatalf("New ran %d times", made)
	}
}

func TestSearchKeepsWhatMatches(t *testing.T) {
	h := New(testVault(t, "Ideas", "Wireframes", "Ideation"))
	h.scan()
	for _, c := range []struct {
		query string
		want  int
	}{
		{"", 3},
		{"idea", 2},
		{"WIRE", 1},
		{"nothing", 0},
	} {
		h.query = c.query
		if got := len(h.shown()); got != c.want {
			t.Errorf("shown(%q) is %d drawings, want %d", c.query, got, c.want)
		}
	}
}

func TestWords(t *testing.T) {
	if got := plural(1, "page"); got != "1 page" {
		t.Errorf("plural(1) is %q", got)
	}
	if got := plural(3, "page"); got != "3 pages" {
		t.Errorf("plural(3) is %q", got)
	}
	if got := humanSize(1536); got != "2 KB" {
		t.Errorf("humanSize(1536) is %q", got)
	}
	if got := ago(time.Now().Add(-90 * time.Minute)); got != "1 h ago" {
		t.Errorf("ago(90m) is %q", got)
	}
	if got := ago(time.Now().Add(-50 * time.Hour)); got != "2 days ago" {
		t.Errorf("ago(50h) is %q", got)
	}
}
