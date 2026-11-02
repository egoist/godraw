// Package vault is the folder godraw keeps its drawings in. The app makes
// it the first time it runs and uses it ever after: every drawing starts
// there, and the home screen shows what it holds.
package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

// Ext is the extension of a drawing.
const Ext = ".tldr"

// Drawing is a file in the vault.
type Drawing struct {
	Path string
	Name string // its file's name, without the extension
	Mod  time.Time
	Size int64
}

// Vault is a folder of drawings.
type Vault struct {
	dir   string
	fresh bool // made now, rather than found
}

// New makes dir the vault, making the folder if it isn't there.
func New(dir string) (*Vault, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("vault: no folder")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	_, statErr := os.Stat(abs)
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	return &Vault{dir: abs, fresh: errors.Is(statErr, os.ErrNotExist)}, nil
}

// Open returns the folder the app keeps its drawings in, making it if it
// isn't there: $GODRAW_VAULT, else the folder the app used last time, else
// ~/Documents/godraw.
func Open() (*Vault, error) {
	if dir := strings.TrimSpace(os.Getenv("GODRAW_VAULT")); dir != "" {
		return New(dir)
	}
	s := loadSettings()
	if s.Vault != "" {
		if v, err := New(s.Vault); err == nil {
			return v, nil
		}
	}
	base, err := mygo.App.Path(mygo.PathDocuments)
	if err != nil || base == "" {
		if base, err = os.UserHomeDir(); err != nil {
			return nil, err
		}
	}
	v, err := New(filepath.Join(base, "godraw"))
	if err != nil {
		return nil, err
	}
	s.Vault = v.dir
	saveSettings(s)
	return v, nil
}

// Dir returns the folder's path.
func (v *Vault) Dir() string { return v.dir }

// Name returns the folder's name, which the home screen shows.
func (v *Vault) Name() string { return filepath.Base(v.dir) }

// Drawings lists the drawings in the folder, the newest first.
func (v *Vault) Drawings() ([]Drawing, error) {
	ents, err := os.ReadDir(v.dir)
	if err != nil {
		return nil, err
	}
	out := make([]Drawing, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") || !strings.EqualFold(filepath.Ext(e.Name()), Ext) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Drawing{
			Path: filepath.Join(v.dir, e.Name()),
			Name: strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
			Mod:  info.ModTime(),
			Size: info.Size(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Mod.Equal(out[j].Mod) {
			return out[i].Mod.After(out[j].Mod)
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// Unique returns a path for a new drawing named name, numbered after the
// name if that one is taken.
func (v *Vault) Unique(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Drawing"
	}
	path := filepath.Join(v.dir, name+Ext)
	for i := 2; ; i++ {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return path
		}
		path = filepath.Join(v.dir, fmt.Sprintf("%s %d%s", name, i, Ext))
	}
}

// Import copies a drawing into the folder, numbered after its name if that
// one is taken, and returns where it went.
func (v *Vault) Import(path string) (string, error) {
	src, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer src.Close()
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	dst := v.Unique(name)
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, src); err != nil {
		out.Close()
		os.Remove(dst)
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	return dst, nil
}

// settings is what the app remembers between launches.
type settings struct {
	Vault string `json:"vault,omitempty"`
}

func settingsFile() string {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "home.json")
}

func loadSettings() settings {
	var s settings
	path := settingsFile()
	if path == "" {
		return s
	}
	if b, err := os.ReadFile(path); err == nil {
		json.Unmarshal(b, &s)
	}
	return s
}

func saveSettings(s settings) {
	path := settingsFile()
	if path == "" {
		return
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, append(b, '\n'), 0o644)
}
