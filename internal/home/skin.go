package home

import "github.com/egoist/mygo/ui"

// skin is the look of the home screen: the theme its widgets take, and the
// few colors of its own that go with it.
type skin struct {
	t    *ui.Theme
	dark bool

	accent, accentDeep, accentSoft ui.Color
	sidebar, card, cardHover       ui.Color
	hairline, preview, chip        ui.Color
}

// skinFor builds the skin of the window, dark or light as asked, else as
// the system has it.
func skinFor(c *ui.Context, mode string) *skin {
	dark := c.Theme().Dark
	switch mode {
	case "light":
		dark = false
	case "dark":
		dark = true
	}
	t := ui.LightTheme()
	if dark {
		t = ui.DarkTheme()
	}
	s := &skin{t: t, dark: dark}
	s.accent = ui.Hex("#6b4df6")
	if dark {
		s.accent = ui.Hex("#8b6cff")
	}
	// The color the user chose on the desktop, where there is one.
	if p := c.Preferences().Accent; p.A != 0 {
		s.accent = p
	}
	s.accentDeep = s.accent.Mix(ui.Hex("#241a4d"), 0.55)
	s.accentSoft = s.accent.Alpha(0.14)

	t.Accent, t.AccentText = s.accent, ui.Hex("#ffffff")
	t.AccentHover = s.accent.Mix(ui.Hex("#ffffff"), 0.14)
	t.AccentPressed = s.accent.Mix(ui.Hex("#000000"), 0.16)
	t.Radius = 8

	if dark {
		t.Background = ui.Hex("#141419")
		t.Surface = ui.Hex("#1d1d25")
		t.SurfaceHover = ui.Hex("#272732")
		t.SurfacePressed = ui.Hex("#31313d")
		t.Border = ui.Hex("#2c2c37")
		t.Text = ui.Hex("#ececf3")
		t.TextMuted = ui.Hex("#8c8c9d")
		s.sidebar = ui.Hex("#181820")
		s.card = ui.Hex("#1b1b23")
		s.cardHover = ui.Hex("#21212b")
		s.preview = ui.Hex("#23232d")
		s.hairline = ui.RGBA(255, 255, 255, 0.07)
		s.chip = ui.RGBA(255, 255, 255, 0.06)
	} else {
		t.Background = ui.Hex("#fbfbfd")
		t.Surface = ui.Hex("#ffffff")
		t.SurfaceHover = ui.Hex("#f1f1f6")
		t.SurfacePressed = ui.Hex("#e6e6ee")
		t.Border = ui.Hex("#e2e2ea")
		t.Text = ui.Hex("#191920")
		t.TextMuted = ui.Hex("#6f6f80")
		s.sidebar = ui.Hex("#f5f5f9")
		s.card = ui.Hex("#ffffff")
		s.cardHover = ui.Hex("#ffffff")
		s.preview = ui.Hex("#eeeef4")
		s.hairline = ui.RGBA(0, 0, 0, 0.08)
		s.chip = ui.RGBA(0, 0, 0, 0.045)
	}
	return s
}

// shadow returns the shadow a raised card casts.
func (s *skin) shadow() ui.Color {
	if s.dark {
		return ui.RGBA(0, 0, 0, 0.55)
	}
	return ui.RGBA(0, 0, 0, 0.12)
}
