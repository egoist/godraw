package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newVault(t *testing.T, name string) (*Vault, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	v, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	return v, dir
}

func TestNewMakesTheFolder(t *testing.T) {
	v, dir := newVault(t, "drawings")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("the folder wasn't made: %v", err)
	}
	if v.Dir() != dir {
		t.Errorf("Dir is %q, want %q", v.Dir(), dir)
	}
	if v.Name() != "drawings" {
		t.Errorf("Name is %q", v.Name())
	}
	if !v.fresh {
		t.Error("a folder made now is fresh")
	}
	again, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	if again.fresh {
		t.Error("a folder found is not fresh")
	}
}

func TestWelcomeIsWrittenOnce(t *testing.T) {
	v, _ := newVault(t, "drawings")
	if err := v.Welcome(); err != nil {
		t.Fatal(err)
	}
	list, err := v.Drawings()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "Welcome" {
		t.Fatalf("got %d drawings: %v", len(list), list)
	}
	// What the user makes of it stays as it is.
	edited := []byte(`{"edited":true}`)
	if err := os.WriteFile(list[0].Path, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := New(v.Dir())
	if err != nil {
		t.Fatal(err)
	}
	if err := again.Welcome(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(list[0].Path); string(b) != string(edited) {
		t.Error("the welcome drawing was written again")
	}
}

func TestDrawingsAreNewestFirst(t *testing.T) {
	v, dir := newVault(t, "drawings")
	now := time.Now()
	for i, name := range []string{"Old", "New", "Middle"} {
		path := filepath.Join(dir, name+".tldr")
		if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		when := now.Add(time.Duration(i-2) * time.Hour)
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
	}
	// Files that are not drawings, and folders, stay out of it.
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0o644)
	os.Mkdir(filepath.Join(dir, "folder.tldr"), 0o755)

	list, err := v.Drawings()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, d := range list {
		names = append(names, d.Name)
	}
	want := []string{"Middle", "New", "Old"}
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("got %v, want %v", names, want)
		}
	}
}

func TestUniqueNumbersTakenNames(t *testing.T) {
	v, dir := newVault(t, "drawings")
	if got, want := v.Unique("Drawing"), filepath.Join(dir, "Drawing.tldr"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	os.WriteFile(filepath.Join(dir, "Drawing.tldr"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(dir, "Drawing 2.tldr"), []byte("{}"), 0o644)
	if got, want := v.Unique("Drawing"), filepath.Join(dir, "Drawing 3.tldr"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := v.Unique("  "), filepath.Join(dir, "Drawing 3.tldr"); got != want {
		t.Fatalf("a blank name got %q, want %q", got, want)
	}
}

func TestImportCopiesIntoTheFolder(t *testing.T) {
	v, dir := newVault(t, "drawings")
	src := filepath.Join(t.TempDir(), "Borrowed.tldr")
	if err := os.WriteFile(src, []byte("tldr"), 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := v.Import(src)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "Borrowed.tldr"); path != want {
		t.Fatalf("got %q, want %q", path, want)
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "tldr" {
		t.Fatalf("the copy is %q, %v", b, err)
	}
	// The file it came from stays where it was.
	if _, err := os.Stat(src); err != nil {
		t.Errorf("the original is gone: %v", err)
	}
	// A second one of that name is numbered.
	if path, err = v.Import(src); err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "Borrowed 2.tldr"); path != want {
		t.Fatalf("got %q, want %q", path, want)
	}
}
