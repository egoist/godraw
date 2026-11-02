// Package home is the window godraw opens on: the folder its drawings are
// kept in, and what is in it. It is the app's front door, as Obsidian's
// vault screen is its.
package home

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"

	"godraw/internal/assets"
	"godraw/internal/vault"
)

// Home is the state of the window: the folder, what it holds, and what the
// user typed to find something in it.
type Home struct {
	v *vault.Vault

	// What the window asks of the app.
	Open   func(path string) // open a drawing in a window of its own
	New    func()            // make a drawing in the folder, and open it
	Reveal func()            // show the folder in the file manager

	theme string // "", "light" or "dark": the system's while empty

	query string

	drawings []vault.Drawing
	scanned  time.Time

	thumbs  map[string]*thumb
	drawn   int  // pictures drawn this frame
	waiting bool // pictures still to draw
}

// New returns the state of the window showing v.
func New(v *vault.Vault) *Home {
	return &Home{v: v, thumbs: map[string]*thumb{}}
}

// Refresh reads the folder again on the next frame.
func (h *Home) Refresh() {
	h.scanned = time.Time{}
}

// View builds the window.
func (h *Home) View(c *ui.Context) {
	s := skinFor(c, h.theme)
	c.SetTheme(s.t)
	h.drawn, h.waiting = 0, false
	h.scan()
	ui.Row(c).Fill().Children(func() {
		h.sidebar(c, s)
		h.main(c, s)
	})
	h.shortcuts(c)
	if h.waiting {
		// Pictures still coming in: draw the next frame.
		c.AnimationFrame()
	}
}

// scan reads the folder again, at most once a second, so that drawings
// saved by another window, or by another app, show up.
func (h *Home) scan() {
	if time.Since(h.scanned) < time.Second {
		return
	}
	h.scanned = time.Now()
	if list, err := h.v.Drawings(); err == nil {
		h.drawings = list
	}
}

// shown returns the drawings the search keeps.
func (h *Home) shown() []vault.Drawing {
	q := strings.ToLower(strings.TrimSpace(h.query))
	if q == "" {
		return h.drawings
	}
	out := make([]vault.Drawing, 0, len(h.drawings))
	for _, d := range h.drawings {
		if strings.Contains(strings.ToLower(d.Name), q) {
			out = append(out, d)
		}
	}
	return out
}

// --- the sidebar ---

func (h *Home) sidebar(c *ui.Context, s *skin) {
	ui.Column(c.Key("sidebar")).Width(268).FillHeight().
		Background(s.sidebar).
		BorderWidth(0, 1, 0, 0).BorderColor(s.hairline).
		Children(func() {
			h.brand(c, s)
			ui.Column(c).Padding(10, 0).Children(func() {
				h.newButton(c, s)
			})
			h.recent(c, s)
			ui.Spacer(c)
			h.footer(c, s)
		})
}

// brand shows the app and the folder it keeps drawings in.
func (h *Home) brand(c *ui.Context, s *skin) {
	ui.Row(c.Key("brand")).Padding(16, 16, 12, 16).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(30, 30).Radius(9).Center().
			LinearGradient(ui.LinearGradient{From: s.accent, To: s.accentDeep, Angle: 145}).
			Shadow(0, 2, 8, 0, s.accent.Alpha(0.35)).
			Children(func() {
				ui.Icon(c, icon("pencil")).Size(16, 16).TextColor(ui.Hex("#ffffff"))
			})
		ui.Column(c).Gap(1).Grow(1).Children(func() {
			ui.Text(c, "godraw").FontSize(13.5).FontWeight(650).TextColor(s.t.Text).SingleLine()
			ui.Text(c, h.v.Name()).FontSize(11.5).TextColor(s.t.TextMuted).SingleLine().
				Tooltip(h.v.Dir()).Cursor(ui.CursorPointer).
				OnClick(func() { h.reveal() })
		})
	})
}

// newButton starts a drawing, the one thing the window is for.
func (h *Home) newButton(c *ui.Context, s *skin) {
	b := ui.ButtonBase(c.Key("new")).Height(36).Radius(9).PaddingX(10).Gap(8).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).Label("New drawing").
		Background(s.accent).TextColor(s.t.AccentText)
	if b.Pressed() {
		b.Background(s.t.AccentPressed)
	} else if b.Hovered() {
		b.Background(s.t.AccentHover)
	}
	b.Children(func() {
		ui.Icon(c, icon("plus")).Size(15, 15)
		ui.Text(c, "New drawing").FontSize(12.5).FontWeight(600).Grow(1).TextColor(s.t.AccentText)
		ui.Text(c, shortcutKey).FontSize(11).TextColor(s.t.AccentText.Alpha(0.65))
	})
	b.OnClick(h.create)
}

// recent lists the drawings touched last.
func (h *Home) recent(c *ui.Context, s *skin) {
	ui.Column(c.Key("recent")).Padding(18, 10, 8, 10).Gap(1).Children(func() {
		ui.Row(c).Padding(6, 4).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, icon("clock")).Size(11, 11).TextColor(s.t.TextMuted.Alpha(0.8))
			ui.Text(c, "RECENT").FontSize(10).FontWeight(700).LetterSpacing(0.09).TextColor(s.t.TextMuted)
		})
		list := h.drawings
		if len(list) > 5 {
			list = list[:5]
		}
		if len(list) == 0 {
			ui.Text(c, "Nothing yet.").FontSize(11.5).TextColor(s.t.TextMuted).Padding(8, 4)
			return
		}
		for _, d := range list {
			h.row(c, s, d)
		}
	})
}

// row is a drawing in the sidebar's list.
func (h *Home) row(c *ui.Context, s *skin, d vault.Drawing) {
	b := ui.ButtonBase(c.Key("row:" + d.Path)).Height(30).Radius(6).PaddingX(8).Gap(8).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).Label(d.Name)
	if b.Hovered() {
		b.Background(s.chip)
	}
	b.Children(func() {
		ui.Icon(c, icon("file-text")).Size(14, 14).TextColor(s.t.TextMuted.Alpha(0.85))
		ui.Text(c, d.Name).FontSize(12).TextColor(s.t.Text).Grow(1).SingleLine()
		ui.Text(c, agoShort(d.Mod)).FontSize(10).TextColor(s.t.TextMuted.Alpha(0.8)).SingleLine()
	})
	b.OnClick(func() { h.open(d.Path) })
}

// footer says what the folder holds, and offers to show it.
func (h *Home) footer(c *ui.Context, s *skin) {
	var size int64
	for _, d := range h.drawings {
		size += d.Size
	}
	ui.Column(c.Key("footer")).Padding(0, 12, 12, 12).Gap(8).Children(func() {
		ui.Box(c).Height(1).FillWidth().Background(s.hairline)
		ui.Row(c).Gap(2).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("%s · %s", plural(len(h.drawings), "drawing"), humanSize(size))).
				FontSize(10.5).TextColor(s.t.TextMuted).Grow(1).SingleLine().
				Tooltip(h.v.Dir())
			h.iconButton(c, s, "folder-open", "Show in folder", func() { h.reveal() })
			name, tip := "moon", "Dark theme"
			if s.dark {
				name, tip = "sun", "Light theme"
			}
			h.iconButton(c, s, name, tip, func() {
				if s.dark {
					h.theme = "light"
				} else {
					h.theme = "dark"
				}
			})
		})
	})
}

// iconButton is a small, quiet button in a row of them.
func (h *Home) iconButton(c *ui.Context, s *skin, name, tip string, fn func()) {
	b := ui.ButtonBase(c.Key("ib:"+name)).Size(26, 26).Radius(6).Center().
		Cursor(ui.CursorPointer).Label(tip).Tooltip(tip)
	if b.Hovered() {
		b.Background(s.chip)
	}
	b.Children(func() { ui.Icon(c, icon(name)).Size(14, 14).TextColor(s.t.TextMuted) })
	b.OnClick(fn)
}

// --- what the window does ---

func (h *Home) open(path string) {
	if h.Open != nil {
		h.Open(path)
	}
}

func (h *Home) create() {
	if h.New != nil {
		h.New()
	}
	h.Refresh()
}

func (h *Home) reveal() {
	if h.Reveal != nil {
		h.Reveal()
	}
}

func (h *Home) shortcuts(c *ui.Context) {
	if c.Shortcut(0, ui.KeyEscape) && h.query != "" {
		h.query = ""
	}
}

// --- icons and words ---

var (
	iconCache = map[string]*ui.SVG{}
	iconMu    sync.Mutex
)

// icon returns one of the icons the app ships, or nil.
func icon(name string) *ui.SVG {
	iconMu.Lock()
	defer iconMu.Unlock()
	if s, ok := iconCache[name]; ok {
		return s
	}
	data, err := assets.Icons.ReadFile("icons/" + name + ".svg")
	if err != nil {
		return nil
	}
	s := ui.MustParseSVG(data)
	iconCache[name] = s
	return s
}

// shortcutKey is what starts a drawing, as the menu shows it.
const shortcutKey = "⌘N"

// ago says when a drawing was last saved, in a few words.
func ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d h ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
	return t.Format("2 Jan 2006")
}

// agoShort says the same in the room a row of the sidebar has.
func agoShort(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	return t.Format("2 Jan")
}

// humanSize says how big a drawing is, as a file manager does.
func humanSize(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.0f KB", float64(n)/1024)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(n)/(1024*1024*1024))
}

// plural counts a noun, in the English the app speaks.
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// shortPath shows a folder from the home folder down, as a shell does.
func shortPath(path string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if strings.HasPrefix(path, home+string(filepath.Separator)) {
			return "~" + path[len(home):]
		}
	}
	return path
}
