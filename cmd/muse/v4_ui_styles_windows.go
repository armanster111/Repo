//go:build windows

package main

import "time"

type uiStyle int

const (
	uiStyleClassic uiStyle = iota
	uiStyleMinimal
	uiStyleCompact
	uiStyleGlass
	uiStyleStudio
	uiStyleLight
	uiStyleCount
)

// uiChromeLayout holds per-frame UI geometry for the active style.
type uiChromeLayout struct {
	marginL      int32
	marginR      int32
	titleY       int32
	statusY      int32
	metaY        int32
	buttonTop    int32
	buttonW      int32
	buttonH      int32
	buttonGap    int32
	vizTop       int32
	vizBottom    int32
	progressTop  int32
	playlistLeft int32
	playlistW    int32
	playlistTop  int32
	showPlaylist bool
	showHints    bool
	showVolume   bool
	panelDim     float64
	drawBorder   bool
}

func (s *appState) cycleUIStyle() {
	s.uiStyle = (s.uiStyle + 1) % uiStyleCount
	s.flashStatus("UI style: "+uiStyleName(s.uiStyle), 2*time.Second)
	s.saveSettings()
	invalidate()
}

func uiStyleName(style uiStyle) string {
	switch style {
	case uiStyleMinimal:
		return "Minimal"
	case uiStyleCompact:
		return "Compact"
	case uiStyleGlass:
		return "Glass"
	case uiStyleStudio:
		return "Studio"
	case uiStyleLight:
		return "Light"
	default:
		return "Classic"
	}
}

func uiStylePreviewColor(style uiStyle) uintptr {
	switch style {
	case uiStyleMinimal:
		return rgb(30, 34, 48)
	case uiStyleCompact:
		return rgb(22, 28, 42)
	case uiStyleGlass:
		return rgb(50, 70, 100)
	case uiStyleStudio:
		return rgb(14, 16, 24)
	case uiStyleLight:
		return rgb(230, 234, 244)
	default:
		return rgb(26, 32, 51)
	}
}

func layoutFor(style uiStyle, width, height int32) uiChromeLayout {
	m := uiChromeLayout{
		marginL: 28, marginR: 28,
		titleY: 24, metaY: 52, statusY: 52,
		buttonTop: 74, buttonW: 78, buttonH: 28, buttonGap: 8,
		vizTop: 112, vizBottom: height - 104,
		progressTop: height - 82,
		playlistLeft: width - 292, playlistW: 264, playlistTop: 128,
		showPlaylist: true, showHints: true, showVolume: true,
		panelDim: 0.75, drawBorder: false,
	}
	switch style {
	case uiStyleMinimal:
		m.marginL, m.marginR = 16, 16
		m.titleY = 18
		m.metaY, m.statusY = 40, 40
		m.buttonTop = 58
		m.buttonW, m.buttonH, m.buttonGap = 70, 24, 6
		m.vizTop = 88
		m.vizBottom = height - 72
		m.progressTop = height - 58
		m.showPlaylist = false
		m.showHints = false
		m.panelDim = 0.65
	case uiStyleCompact:
		m.buttonW, m.buttonH, m.buttonGap = 60, 22, 4
		m.buttonTop = 68
		m.vizTop = 98
		m.vizBottom = height - 88
		m.progressTop = height - 74
		m.playlistTop = 118
		m.panelDim = 0.8
	case uiStyleGlass:
		m.panelDim = 0.48
		m.drawBorder = true
	case uiStyleStudio:
		m.playlistLeft = 28
		m.playlistW = 248
		m.playlistTop = 112
		m.buttonTop = height - 118
		m.vizTop = 112
		m.vizBottom = height - 132
		m.progressTop = height - 96
		m.panelDim = 0.82
		m.drawBorder = true
	case uiStyleLight:
		m.panelDim = 0.88
		m.drawBorder = true
	}
	if m.playlistLeft+m.playlistW > width-m.marginR {
		m.showPlaylist = false
	}
	return m
}

func uiPaletteForStyle(style uiStyle, base palette) palette {
	if style != uiStyleLight {
		return base
	}
	return palette{
		background: rgb(245, 247, 252),
		panel:      rgb(214, 220, 234),
		text:       rgb(22, 26, 36),
		dim:        rgb(88, 98, 118),
		accent:     rgb(36, 118, 220),
		accent2:    rgb(118, 58, 210),
	}
}

func fillPanel(hdc uintptr, r rect, p palette, chrome uiChromeLayout) {
	fill(hdc, r, dimColor(p.panel, chrome.panelDim))
	if chrome.drawBorder {
		drawPanelBorder(hdc, r, dimColor(p.accent2, 0.85), 1)
	}
}

func drawPanelBorder(hdc uintptr, r rect, col uintptr, thickness int32) {
	line(hdc, r.left, r.top, r.right, r.top, col, thickness)
	line(hdc, r.left, r.bottom, r.right, r.bottom, col, thickness)
	line(hdc, r.left, r.top, r.left, r.bottom, col, thickness)
	line(hdc, r.right, r.top, r.right, r.bottom, col, thickness)
}

func (chrome uiChromeLayout) visualBounds(width, height int32) rect {
	left := chrome.marginL
	right := width - chrome.marginR
	if chrome.showPlaylist && app.uiStyle == uiStyleStudio {
		left = chrome.marginL + chrome.playlistW + 12
	}
	top := chrome.vizTop
	bottom := chrome.vizBottom
	if bottom <= top {
		bottom = top + 32
	}
	return rect{left: left, top: top, right: right, bottom: bottom}
}

func (chrome uiChromeLayout) progressBar(width int32) rect {
	return rect{
		left:   chrome.marginL,
		top:    chrome.progressTop,
		right:  width - chrome.marginR,
		bottom: chrome.progressTop + 14,
	}
}

func volumeSliderRectFor(chrome uiChromeLayout, width int32) rect {
	return rect{
		left:   width - 260,
		top:    chrome.buttonTop + 2,
		right:  width - 40,
		bottom: chrome.buttonTop + 16,
	}
}

func (s *appState) drawUIStylePicker(hdc uintptr, left, top int32, palette palette, chrome uiChromeLayout) {
	s.uiStylePreviewBounds = s.uiStylePreviewBounds[:0]
	procSetTextColor.Call(hdc, palette.dim)
	textOut(hdc, left, top, "UI style (click) — F8 to cycle:")
	box := int32(36)
	gap := int32(8)
	for i := uiStyle(0); i < uiStyleCount; i++ {
		idx := int(i)
		x := left + int32(idx)*(box+gap)
		y := top + 18
		r := rect{left: x, top: y, right: x + box, bottom: y + 22}
		s.uiStylePreviewBounds = append(s.uiStylePreviewBounds, r)
		col := uiStylePreviewColor(i)
		if i == s.uiStyle {
			fill(hdc, r, col)
			drawPanelBorder(hdc, r, palette.accent2, 2)
		} else {
			fill(hdc, r, dimColor(col, 0.6))
		}
		procSetTextColor.Call(hdc, palette.text)
		label := uiStyleName(i)
		if len(label) > 5 {
			label = label[:5]
		}
		textOut(hdc, x+4, y+6, label)
	}
}

func (s *appState) handleUIStyleClick(x, y int32) bool {
	if !s.showSettings {
		return false
	}
	for i, r := range s.uiStylePreviewBounds {
		if pointInRect(x, y, r) {
			s.uiStyle = uiStyle(i % int(uiStyleCount))
			s.flashStatus("UI style: "+uiStyleName(s.uiStyle), 2*time.Second)
			s.saveSettings()
			invalidate()
			return true
		}
	}
	return false
}
