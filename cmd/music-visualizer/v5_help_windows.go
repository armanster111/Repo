//go:build windows

package main

func (s *appState) toggleHelp() {
	s.showHelp = !s.showHelp
	if s.showHelp {
		s.showSettings = false
	}
	invalidate()
}

func (s *appState) drawHelpPanel(hdc uintptr, width, height int32, palette palette, chrome uiChromeLayout) {
	if !s.showHelp {
		return
	}
	left := chrome.marginL + 40
	top := chrome.vizTop
	right := width - chrome.marginR - 40
	bottom := chrome.vizBottom
	if bottom-top < 200 {
		bottom = top + 200
	}
	panel := rect{left: left, top: top, right: right, bottom: bottom}
	fillPanel(hdc, panel, palette, uiChromeLayout{panelDim: 0.94, drawBorder: true})
	procSetTextColor.Call(hdc, palette.text)
	textOut(hdc, left+14, top+12, "Keyboard shortcuts (F1 or ? to close)")
	procSetTextColor.Call(hdc, palette.dim)
	lines := []string{
		"Playback:  Space play/pause  N next  B prev  R restart  M mute",
		"           Q repeat  P shuffle  J speed  , . DJ crossfader",
		"Visuals:   V cycle mode  G theme  F8 UI style  F9 cinema  Z visual-only",
		"           F11 fullscreen  X mini  0 auto-preset  6 party mode",
		"Export:    C PNG snapshot  2 GIF  * AVI video",
		"Library:   O open  D folder  7 scan  \\ panel  U recent  Y favorites",
		"Stream:    1 OBS overlay  [ ] overlay opacity  Esc layered exit",
		"Panels:    3 settings  I desktop  H save theme  8/9 presets",
		"Playlist:  wheel or PgUp/PgDn scroll list when library open",
	}
	y := top + 36
	for _, line := range lines {
		textOut(hdc, left+14, y, line)
		y += 18
	}
}
