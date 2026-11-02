package home

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"godraw/internal/vault"
)

// main is the right of the window: what the folder holds.
func (h *Home) main(c *ui.Context, s *skin) {
	ui.Column(c.Key("main")).Grow(1).FillHeight().Children(func() {
		h.head(c, s)
		ui.Column(c).Grow(1).Padding(0, 28, 20, 28).Gap(18).Children(func() {
			if len(h.drawings) == 0 {
				h.firstRun(c, s)
				return
			}
			h.hero(c, s)
			h.grid(c, s)
		})
	})
}

// head says where the drawings are, and finds one.
func (h *Home) head(c *ui.Context, s *skin) {
	ui.Row(c.Key("head")).Padding(26, 28, 16, 28).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Gap(3).Grow(1).Children(func() {
			ui.Text(c, "Drawings").FontSize(20).FontWeight(700).TextColor(s.t.Text)
			ui.Text(c, "Everything in "+shortPath(h.v.Dir())).FontSize(11.5).
				TextColor(s.t.TextMuted).SingleLine()
		})
		ui.SearchField(c, &h.query).Width(232).Placeholder("Find a drawing")
	})
}

// hero starts a drawing, in the color of the app.
func (h *Home) hero(c *ui.Context, s *skin) {
	white := ui.Hex("#ffffff")
	box := ui.Box(c.Key("hero")).Height(124).Radius(14).Clip().
		Gradient(s.accent, s.accentDeep, 118)
	box.Children(func() {
		ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().
			RadialGradient(ui.RadialGradient{
				Center: ui.Point{X: 0.8, Y: 0.05},
				Stops: []ui.GradientStop{
					{Color: white.Alpha(0.20), At: 0},
					{Color: white.Alpha(0), At: 1},
				},
			})
		// A few shapes, as a page of a drawing would have.
		ui.Box(c).Absolute().Size(128, 84).Top(-26).Right(84).Rotate(14).
			Radius(12).Border(1.5, white.Alpha(0.22)).PassThrough()
		ui.Box(c).Absolute().Size(104, 104).Top(34).Right(-22).Rotate(-12).
			Radius(14).Border(1.5, white.Alpha(0.16)).PassThrough()
		ui.Box(c).Absolute().Size(30, 30).Top(74).Right(196).Rotate(0).
			Radius(15).Border(1.5, white.Alpha(0.3)).PassThrough()
		ui.Row(c).Fill().Padding(22, 22).Gap(20).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Gap(7).Grow(1).Children(func() {
				ui.Text(c, "Start something new").FontSize(17).FontWeight(700).TextColor(white)
				ui.Text(c, "A fresh page, saved straight into your folder.").FontSize(12).
					TextColor(white.Alpha(0.75)).SingleLine()
				ui.Row(c).Children(func() {
					b := ui.ButtonBase(c.Key("hero-new")).Height(31).Radius(8).PaddingX(12).Gap(7).
						AlignItems(ui.Center).Background(white).Cursor(ui.CursorPointer).Label("New drawing")
					if b.Pressed() {
						b.Background(white.Alpha(0.85))
					} else if b.Hovered() {
						b.Background(white.Alpha(0.93))
					}
					b.Children(func() {
						ui.Icon(c, icon("plus")).Size(14, 14).TextColor(s.accent)
						ui.Text(c, "New drawing").FontSize(12).FontWeight(650).TextColor(s.accent)
					})
					b.OnClick(h.create)
				})
			})
		})
	})
}

// grid shows every drawing in the folder, newest first.
func (h *Home) grid(c *ui.Context, s *skin) {
	list := h.shown()
	ui.Column(c.Key("grid")).Grow(1).Gap(10).Children(func() {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "ALL DRAWINGS").FontSize(10).FontWeight(700).LetterSpacing(0.09).
				TextColor(s.t.TextMuted)
			if q := strings.TrimSpace(h.query); q != "" {
				ui.Text(c, fmt.Sprintf("%d of %d", len(list), len(h.drawings))).FontSize(10.5).
					TextColor(s.t.TextMuted)
			}
			ui.Spacer(c)
			ui.Text(c, "Newest first").FontSize(10.5).TextColor(s.t.TextMuted.Alpha(0.8))
		})
		if len(list) == 0 {
			h.noMatch(c, s)
			return
		}
		ui.GridView(c, nil, len(list), 236, 196, func(i int) { h.card(c, s, list[i]) }).Grow(1)
	})
}

// card is one drawing: a picture of its page, its name and what is on it.
func (h *Home) card(c *ui.Context, s *skin, d vault.Drawing) {
	t := h.thumbOf(d, s.dark)
	b := ui.ButtonBase(c.Key("card:"+d.Path)).Column().Grow(1).FillWidth().
		Radius(11).Padding(6).
		Background(s.card).Border(1, s.hairline).Cursor(ui.CursorPointer).Label(d.Name)
	hover := b.Hovered()
	if hover {
		b.BorderColor(s.accent.Alpha(0.55)).Background(s.cardHover).
			Shadow(0, 8, 22, 0, s.shadow())
	}
	b.Children(func() {
		shot := ui.Box(c).Grow(1).FillWidth().Radius(7).Clip().Background(s.preview)
		switch {
		case t != nil && t.bmp != nil:
			shot.Draw(func(p *ui.Painter, r ui.Rect) { drawThumb(p, r, t.bmp) })
		case t != nil && t.read && t.shapes == 0:
			shot.Children(func() { emptyPage(c, s) })
		}
		ui.Row(c).FillWidth().Padding(8, 8, 4, 8).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Gap(1).Grow(1).AlignItems(ui.Start).Children(func() {
				ui.Text(c, d.Name).FontSize(12.5).FontWeight(600).TextColor(s.t.Text).SingleLine()
				ui.Text(c, metaOf(t, d)).FontSize(10.5).TextColor(s.t.TextMuted).SingleLine()
			})
			if hover {
				ui.Icon(c, icon("arrow-up-right")).Size(13, 13).TextColor(s.t.TextMuted)
			}
		})
	})
	b.OnClick(func() { h.open(d.Path) })
}

// drawThumb draws a picture of a drawing in a box, whole and in the
// middle, as large as it fits.
func drawThumb(p *ui.Painter, r ui.Rect, bmp *ui.Bitmap) {
	iw, ih := bmp.Size()
	if iw <= 0 || ih <= 0 {
		return
	}
	w, h := float32(iw), float32(ih)
	k := min(r.W/w, r.H/h)
	dw, dh := w*k, h*k
	p.ImageRegion(bmp, ui.Rect{}, ui.Rect{
		X: r.X + (r.W-dw)/2, Y: r.Y + (r.H-dh)/2, W: dw, H: dh,
	}, ui.ImageOptions{})
}

// metaOf says what a drawing holds, once it has been read.
func metaOf(t *thumb, d vault.Drawing) string {
	if t == nil || !t.read {
		return ago(d.Mod)
	}
	return fmt.Sprintf("%s · %s", plural(t.pages, "page"), ago(d.Mod))
}

// emptyPage is what a drawing with nothing on it shows.
func emptyPage(c *ui.Context, s *skin) {
	ui.Column(c).Fill().Center().Gap(5).Children(func() {
		ui.Icon(c, icon("sparkles")).Size(18, 18).TextColor(s.t.TextMuted.Alpha(0.55))
		ui.Text(c, "Empty page").FontSize(10.5).TextColor(s.t.TextMuted.Alpha(0.75))
	})
}

// firstRun is the window of a folder with nothing in it yet.
func (h *Home) firstRun(c *ui.Context, s *skin) {
	ui.Column(c.Key("first")).Grow(1).FillWidth().Center().Gap(14).Radius(16).
		BorderWidth(1.5).BorderColor(s.t.Border).BorderStyle(ui.BorderDashed).
		Children(func() {
			ui.Box(c).Size(52, 52).Radius(26).Center().Background(s.accentSoft).
				Border(1, s.accent.Alpha(0.25)).Children(func() {
				ui.Icon(c, icon("plus")).Size(22, 22).TextColor(s.accent)
			})
			ui.Column(c).Gap(5).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Your folder is empty").FontSize(15).FontWeight(650).TextColor(s.t.Text)
				ui.Text(c, "Start a drawing, or drop a .tldr file anywhere in this window.").
					FontSize(12).TextColor(s.t.TextMuted)
			})
			ui.PrimaryButton(c, "New drawing").OnClick(h.create)
		})
}

// noMatch is what a search that finds nothing shows.
func (h *Home) noMatch(c *ui.Context, s *skin) {
	ui.Column(c).Grow(1).Center().Gap(8).Children(func() {
		ui.Text(c, "No drawing matches “"+strings.TrimSpace(h.query)+"”").
			FontSize(13).TextColor(s.t.TextMuted)
		ui.Button(c, "Clear the search").OnClick(func() { h.query = "" })
	})
}
