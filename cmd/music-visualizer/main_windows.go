//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/armanster111/music-visualizer/internal/library"
	"github.com/armanster111/music-visualizer/internal/lyrics"
	"github.com/armanster111/music-visualizer/internal/metadata"
	"github.com/armanster111/music-visualizer/internal/presets"
	"github.com/armanster111/music-visualizer/internal/visual"
)

const (
	appTitle = "Music Visualizer"
	fps      = 60
	barCount = 64

	cwUseDefault       = 0x80000000
	wsOverlappedWindow = 0x00cf0000

	swShowDefault   = 10
	swShowNormal    = 1
	swShowMaximized = 3

	wmDestroy      = 0x0002
	wmEraseBkgnd   = 0x0014
	wmPaint        = 0x000f
	wmTimer        = 0x0113
	wmKeyDown     = 0x0100
	wmChar        = 0x0102
	wmCreate      = 0x0001
	wmMouseMove   = 0x0200
	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202
	wmLButtonDblClk = 0x0203
	wmMove          = 0x0003
	wmSize          = 0x0005
	wmDropFiles     = 0x0233
	wmNCHitTest   = 0x0084
	wmApp         = 0x8000
	wmMciNotify   = wmApp + 1

	mkLButton = 0x0001

	vkEscape         = 0x1b
	vkEnd            = 0x23
	vkHome           = 0x24
	vkLeft           = 0x25
	vkUp             = 0x26
	vkRight          = 0x27
	vkDown           = 0x28
	vkSpace          = 0x20
	vkF5             = 0x74
	vkF6             = 0x75
	vkF7             = 0x76
	vkF8             = 0x77
	vkF9             = 0x78
	vkF10            = 0x79
	vkF11            = 0x7a
	vkMediaNext      = 0xb0
	vkMediaPrev      = 0xb1
	vkMediaStop      = 0xb2
	vkMediaPlayPause = 0xb3

	idcArrow = 32512

	timerID = 1

	transparent = 1
	srccopy     = 0x00CC0020

	ofnPathMustExist = 0x00000800
	ofnFileMustExist = 0x00001000
	ofnExplorer      = 0x00080000

	mciNotifySuccess = 0x0001
	mciNotifyAborted = 0x0004

	mbIconError = 0x00000010
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	winmm    = syscall.NewLazyDLL("winmm.dll")

	procBeginPaint          = user32.NewProc("BeginPaint")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procEndPaint            = user32.NewProc("EndPaint")
	procFillRect            = user32.NewProc("FillRect")
	procGetClientRect       = user32.NewProc("GetClientRect")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procInvalidateRect      = user32.NewProc("InvalidateRect")
	procLoadCursorW         = user32.NewProc("LoadCursorW")
	procMessageBoxW         = user32.NewProc("MessageBoxW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procReleaseCapture      = user32.NewProc("ReleaseCapture")
	procSetCapture          = user32.NewProc("SetCapture")
	procSetTimer            = user32.NewProc("SetTimer")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procLoadIconW           = user32.NewProc("LoadIconW")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procShowWindow          = user32.NewProc("ShowWindow")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procUpdateWindow        = user32.NewProc("UpdateWindow")

	procBitBlt               = gdi32.NewProc("BitBlt")
	procSetDIBitsToDevice    = gdi32.NewProc("SetDIBitsToDevice")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	procCreatePen        = gdi32.NewProc("CreatePen")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procDeleteDC         = gdi32.NewProc("DeleteDC")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procEllipse          = gdi32.NewProc("Ellipse")
	procLineTo           = gdi32.NewProc("LineTo")
	procMoveToEx         = gdi32.NewProc("MoveToEx")
	procSelectObject     = gdi32.NewProc("SelectObject")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procTextOutW         = gdi32.NewProc("TextOutW")

	procGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
	procGetOpenFileNameW   = comdlg32.NewProc("GetOpenFileNameW")
	procDragAcceptFiles    = shell32.NewProc("DragAcceptFiles")
	procDragFinish         = shell32.NewProc("DragFinish")
	procDragQueryFileW     = shell32.NewProc("DragQueryFileW")
	procMCIGetErrorStringW = winmm.NewProc("mciGetErrorStringW")
	procMCISendStringW     = winmm.NewProc("mciSendStringW")
)

type visualMode int

const (
	modeClassic visualMode = iota
	modeMirror
	modeBlocks
	modeWave
	modeHalo
	modeSpectrum
	modeFire
	modeParticles
	modeTunnel
	modePlasma
	modeAurora
	modeMandala
	modeStarfield
	modeKaleidoscope
	modeNeonCity
	modeSupernova
	modeLiquid
	modeOrbit
	modeWaveform3D
	modeFluid
	modeGalaxy
	modeChromatic
	modeSpectrogram
	modeOscilloscope
	modeLissajous
	modeVortex
	modeComets
	modeBassDrop
	modeRain
	modePrism
	modeCount
)

type theme int

const (
	themeNeon theme = iota
	themeLava
	themeCyberpunk
	themeOcean
	themeCustomBase
	themeCount = themeCustomBase + 5
)

type repeatMode int

const (
	repeatOff repeatMode = iota
	repeatOne
	repeatAll
)

type settings struct {
	Mode               visualMode    `json:"mode"`
	Theme              theme         `json:"theme"`
	Volume             int           `json:"volume"`
	Muted              bool          `json:"muted"`
	DesktopInput       bool          `json:"desktop_input"`
	DesktopAuto        bool          `json:"desktop_auto"`
	DesktopSensitivity float64       `json:"desktop_sensitivity"`
	DesktopDevice      int           `json:"desktop_device"`
	Repeat             repeatMode    `json:"repeat"`
	Shuffle            bool          `json:"shuffle"`
	Recent             []string      `json:"recent"`
	Favorites          []string      `json:"favorites"`
	LibraryRoot        string        `json:"library_root"`
	SeenOnboarding     bool          `json:"seen_onboarding"`
	WindowX            int32         `json:"window_x"`
	WindowY            int32         `json:"window_y"`
	WindowW            int32         `json:"window_w"`
	WindowH            int32         `json:"window_h"`
	PlaybackSpeed      int           `json:"playback_speed"`
	EqBass             int           `json:"eq_bass"`
	EqMid              int           `json:"eq_mid"`
	EqTreble           int           `json:"eq_treble"`
	CustomThemes       []customTheme `json:"custom_themes"`
	Presets            []presetData  `json:"presets"`
	PresetIndex        int           `json:"preset_index"`
	PlayCounts         map[string]int `json:"play_counts"`
	LastPlayed         map[string]int64 `json:"last_played"`
	LibraryPaths       []string      `json:"library_paths"`
	Crossfade          bool          `json:"crossfade"`
	KaraokeMode        bool          `json:"karaoke_mode"`
	AmbientMode        bool          `json:"ambient_mode"`
	OverlayMode        bool          `json:"overlay_mode"`
	OverlayX           int32         `json:"overlay_x"`
	OverlayY           int32         `json:"overlay_y"`
	OverlayW           int32         `json:"overlay_w"`
	OverlayH           int32         `json:"overlay_h"`
	OverlayAlpha       int           `json:"overlay_alpha"`
	RemoteControl      bool          `json:"remote_control"`
	AutoPreset         bool          `json:"auto_preset"`
	DJMode             bool          `json:"dj_mode"`
	VisualIntensity    float64       `json:"visual_intensity"`
	UIStyle            int           `json:"ui_style"`
	PanelGroup         string        `json:"panel_group"`
	PanelGroupKey      string        `json:"panel_group_key"`
	SettingsVersion    int           `json:"settings_version"`
	LiveFFT            bool          `json:"live_fft"`
	MoodReactive       bool          `json:"mood_reactive"`
	InputSource        int           `json:"input_source"`
	OverlayAspect      int           `json:"overlay_aspect"`
	ChromaKey          bool          `json:"chroma_key"`
}

type track struct {
	Path     string
	Title    string
	Favorite bool
}

type button struct {
	label  string
	action string
	bounds rect
}

type presetData struct {
	Name               string  `json:"name"`
	Mode               int     `json:"mode"`
	Theme              int     `json:"theme"`
	EqBass             int     `json:"eq_bass"`
	EqMid              int     `json:"eq_mid"`
	EqTreble           int     `json:"eq_treble"`
	DesktopSensitivity float64 `json:"desktop_sensitivity"`
	VisualIntensity    float64 `json:"visual_intensity"`
	Karaoke            bool    `json:"karaoke"`
	Ambient            bool    `json:"ambient"`
}

type appState struct {
	hwnd               uintptr
	filePath           string
	status             string
	frames             []visual.Frame
	currentBars        []float64
	ambientSeed        float64
	duration           time.Duration
	startedAt          time.Time
	playbackOffset     time.Duration
	dragPosition       time.Duration
	mode               visualMode
	theme              theme
	playlist           []track
	currentIndex       int
	recent             []string
	favorites          map[string]bool
	buttons            []button
	desktop            *desktopInput
	libraryRoot        string
	volume             int
	sleepMinutes       int
	sleepStartedAt     time.Time
	repeat             repeatMode
	seekLarge          bool
	shuffle            bool
	muted              bool
	desktopMode        bool
	fullscreen         bool
	mini               bool
	playing            bool
	paused             bool
	dragging           bool
	panel              panelView
	searchQuery        string
	visualOnly         bool
	seenOnboarding     bool
	windowX            int32
	windowY            int32
	windowW            int32
	windowH            int32
	playbackSpeed      int
	eqBass             int
	eqMid              int
	eqTreble           int
	eqFocus            int
	desktopSensitivity float64
	desktopAuto        bool
	desktopDevice      int
	customThemes       []customTheme
	meta               trackMeta
	lyricLines         []lyrics.Line
	artGrid            *metadata.ArtGrid
	energyHistory      []float64
	modeBlend          float64
	prevMode           visualMode
	libraryIndex       *library.Index
	libraryScanning    bool
	presetIndex        int
	presets            []presets.Preset
	overlayHWND        uintptr
	overlayMode        bool
	ambientMode        bool
	karaokeMode        bool
	showSettings       bool
	crossfade          bool
	crossfadeLevel     float64
	bpm                float64
	mood               string
	djMode             bool
	djCrossfader       float64
	djDeckBPath        string
	djDeckBFrames      []visual.Frame
	remoteEnabled      bool
	autoPreset         bool
	visualIntensity    float64
	uiStyle            uiStyle
	uiLayout           uiChromeLayout
	uiStylePreviewBounds []rect
	playCounts         map[string]int
	lastPlayed         map[string]int64
	libraryPaths       []string
	panelGroup         string
	panelGroupKey      string
	panelRow           int
	volumeSliderBounds rect
	playlistRowBounds  []rect
	gifRecording       bool
	gifFrameCount      int
	lyricsFetching     bool
	ultra              *ultraEngine
	lastMouseMove      time.Time
	cinemaHideUI       bool
	overlayAlpha       byte
	overlayX           int32
	overlayY           int32
	overlayW           int32
	overlayH           int32
	statusFlash        string
	statusFlashUntil   time.Time
	settingsRows       []settingsRow
	partyMode          bool
	shader             *shaderEngine
	gifFrames          []*visual.Canvas
	mp4Recording       bool
	mp4FrameCount      int
	videoFrames        []*visual.Canvas
	preloadPath        string
	preloadFrames      []visual.Frame
	preloadBusy        bool
	intensitySlider    rect
	modePreviewBounds  []rect
	pcmCache           *visual.PCMCache
	liveFFT            bool
	moodReactive       bool
	inputSource        inputSource
	specHistory        [][]float64
	overlayAspect      overlayAspect
	chromaKey          bool
	onset              *visual.OnsetEngine
	perf               *perfEngine
}

var app = &appState{
	status:             "Open or drag audio here. 30 visualizers, live browser viz at :8765/viz, kick/snare engine ready.",
	currentIndex:       -1,
	theme:              themeNeon,
	volume:             800,
	playbackSpeed:      1000,
	desktopSensitivity: 1.0,
	visualIntensity:    1.0,
	favorites:          map[string]bool{},
	playCounts:         map[string]int{},
	lastPlayed:         map[string]int64{},
	djCrossfader:       0.5,
	crossfadeLevel:     1.0,
	overlayAlpha:       220,
	lastMouseMove:      time.Now(),
	remoteEnabled:      true,
	liveFFT:            true,
	moodReactive:       true,
	desktop:            newDesktopInput(),
	ultra:              newUltraEngine(),
	onset:              visual.NewOnsetEngine(),
	perf:               newPerfEngine(),
}

type point struct {
	x int32
	y int32
}

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type paintStruct struct {
	hdc         uintptr
	erase       int32
	rcPaint     rect
	restore     int32
	incUpdate   int32
	rgbReserved [32]byte
}

type wndClassEx struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   uintptr
	icon       uintptr
	cursor     uintptr
	background uintptr
	menuName   *uint16
	className  *uint16
	iconSm     uintptr
}

type openFileName struct {
	structSize     uint32
	hwndOwner      uintptr
	instance       uintptr
	filter         *uint16
	customFilter   *uint16
	maxCustFilter  uint32
	filterIndex    uint32
	file           *uint16
	maxFile        uint32
	fileTitle      *uint16
	maxFileTitle   uint32
	initialDir     *uint16
	title          *uint16
	flags          uint32
	fileOffset     uint16
	fileExtension  uint16
	defaultExt     *uint16
	custData       uintptr
	hook           uintptr
	templateName   *uint16
	reserved       uintptr
	reservedFlags  uint32
	reservedFlags2 uint32
}

func main() {
	if err := run(); err != nil {
		showMessage(0, appTitle, err.Error(), mbIconError)
	}
}

func run() error {
	runtime.LockOSThread()
	app.loadSettings()
	app.initV2Features()

	instance, _, _ := procGetModuleHandleW.Call(0)
	className, _ := syscall.UTF16PtrFromString("MusicVisualizerWindow")
	title, _ := syscall.UTF16PtrFromString(appTitle)
	cursor, _, _ := procLoadCursorW.Call(0, idcArrow)

	wc := wndClassEx{
		size:      uint32(unsafe.Sizeof(wndClassEx{})),
		wndProc:   syscall.NewCallback(wndProc),
		instance:  instance,
		cursor:    cursor,
		className: className,
	}
	if ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 {
		return fmt.Errorf("register window class: %w", err)
	}

	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		wsOverlappedWindow,
		cwUseDefault,
		cwUseDefault,
		1000,
		640,
		0,
		0,
		instance,
		0,
	)
	if hwnd == 0 {
		return fmt.Errorf("create window: %w", err)
	}
	app.hwnd = hwnd
	procDragAcceptFiles.Call(hwnd, 1)
	app.restoreWindowBounds()
	app.initTray()
	app.showOnboardingIfNeeded()
	if app.desktop != nil {
		app.desktop.setSensitivity(app.desktopSensitivity)
		_ = app.desktop.setDeviceIndex(app.desktopDevice)
	}
	if app.desktopMode {
		_ = app.desktop.start()
	}
	app.restoreOverlayIfNeeded()

	procShowWindow.Call(hwnd, swShowDefault)
	procUpdateWindow.Call(hwnd)

	if len(os.Args) > 1 {
		app.openPath(os.Args[1])
	}

	var message msg
	for {
		ret, _, err := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) == -1 {
			return fmt.Errorf("get message: %w", err)
		}
		if ret == 0 {
			return nil
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func wndProc(hwnd uintptr, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		procSetTimer.Call(hwnd, timerID, 1000/fps, 0)
		return 0
	case wmTimer:
		app.tickSleepTimer()
		app.maybeAutoDesktop()
		app.tickCrossfade()
		app.applyCrossfadeVolume()
		app.tickGaplessPreload()
		app.captureGifFrame()
		app.captureMp4Frame()
		if app.ultra != nil {
			app.ultra.tickShowcase()
		}
		app.tickCinemaUI()
		app.tickStatusFlash()
		if app.modeBlend < 1 {
			app.modeBlend += 0.25
			if app.modeBlend > 1 {
				app.modeBlend = 1
			}
		}
		procInvalidateRect.Call(hwnd, 0, 0)
		if app.overlayHWND != 0 && app.overlayMode {
			procInvalidateRect.Call(app.overlayHWND, 0, 0)
		}
		return 0
	case wmEraseBkgnd:
		return 1
	case wmKeyDown:
		switch wParam {
		case vkEscape:
			app.handleEscapeKey()
		case vkEnd:
			app.nextTrack()
		case vkHome:
			app.seekTo(0, app.playing)
		case vkLeft:
			app.seekBy(-app.seekStep())
		case vkRight:
			app.seekBy(app.seekStep())
		case vkUp:
			app.changeVolume(50)
		case vkDown:
			app.changeVolume(-50)
		case vkSpace:
			app.togglePlay()
		case vkF11:
			app.toggleFullscreen()
		case vkF8:
			app.cycleUIStyle()
		case vkF5:
			app.toggleLiveFFT()
		case vkF6:
			app.toggleMoodReactive()
		case vkF7:
			app.toggleChromaKey()
		case vkF10:
			app.cycleOverlayAspect()
		case vkF9:
			if app.ultra != nil {
				app.ultra.toggleShowcase()
			}
		case vkMediaNext:
			app.nextTrack()
		case vkMediaPrev:
			app.previousTrack()
		case vkMediaStop:
			if app.playing {
				app.togglePlay()
			}
		case vkMediaPlayPause:
			app.togglePlay()
		case uintptr('B'):
			app.previousTrack()
		case uintptr('C'):
			app.exportSnapshot()
		case uintptr('D'):
			app.loadCurrentFolder()
		case uintptr('E'):
			app.focusNextEQBand()
		case uintptr('F'):
			app.toggleFavorite()
		case uintptr('G'):
			app.nextTheme()
		case uintptr('H'):
			app.saveCustomTheme()
		case uintptr('I'):
			app.cycleInputSource()
		case uintptr('J'):
			app.cyclePlaybackSpeed()
		case uintptr('K'):
			app.cycleEQBand()
		case uintptr('L'):
			app.toggleFullscreen()
		case uintptr('M'):
			app.toggleMute()
		case uintptr('N'):
			app.nextTrack()
		case uintptr('O'):
			if path, ok := openWAVDialog(hwnd); ok {
				app.openPath(path)
			}
		case uintptr('P'):
			app.toggleShuffle()
		case uintptr('Q'):
			app.cycleRepeat()
		case uintptr('R'):
			app.restart()
		case uintptr('S'):
			app.cycleSleepTimer()
		case uintptr('T'):
			app.seekLarge = !app.seekLarge
			app.updateStatus()
			invalidate()
		case uintptr('U'):
			app.showRecentView()
		case uintptr('V'):
			app.cycleMode()
		case uintptr('W'):
			app.cycleDesktopDevice()
		case uintptr('X'):
			app.toggleMini()
		case uintptr('Y'):
			app.showFavoritesView()
		case uintptr('Z'):
			app.toggleVisualOnly()
		case uintptr('A'):
			app.toggleDesktopAuto()
		case uintptr('/'):
			app.showSearchView()
		case uintptr(']'):
			if app.overlayMode {
				app.cycleOverlayOpacity(15)
			} else {
				app.cycleDesktopSensitivity()
			}
		case uintptr('['):
			if app.overlayMode {
				app.cycleOverlayOpacity(-15)
			}
		case uintptr('1'):
			app.toggleOverlay()
		case uintptr('2'):
			app.toggleGifRecord()
		case uintptr('*'):
			app.toggleMp4Record()
		case uintptr('3'):
			app.toggleSettingsPanel()
		case uintptr('4'):
			app.exportSyncBundle()
		case uintptr('5'):
			app.importPresets()
		case uintptr('6'):
			app.togglePartyMode()
		case uintptr('7'):
			app.scanLibraryAsync()
		case uintptr('8'):
			app.cyclePreset()
		case uintptr('9'):
			app.saveCurrentPreset()
		case uintptr(','):
			app.djNudgeCrossfader(-0.1)
		case uintptr('.'):
			app.djNudgeCrossfader(0.1)
		case uintptr(';'):
			app.loadDJDeckB()
		case uintptr('\\'):
			app.cycleLibraryPanel()
		case uintptr('='):
			app.toggleCrossfade()
		case uintptr(')'):
			app.toggleKaraoke()
		case uintptr('`'):
			app.toggleAmbient()
		case uintptr('~'):
			app.toggleDJMode()
		case uintptr('0'):
			app.toggleAutoPreset()
		}
		return 0
	case wmChar:
		if app.panel == viewSearch {
			app.appendSearchChar(rune(wParam))
		}
		return 0
	case wmTrayIcon:
		app.handleTrayMessage(lParam)
		return 0
	case wmMciNotify:
		if wParam == mciNotifySuccess {
			app.handleTrackEnded()
		}
		return 0
	case wmDropFiles:
		app.handleDrop(wParam)
		return 0
	case wmLButtonDown:
		x, y := mousePoint(lParam)
		app.noteMouseActivity()
		if app.handleSettingsClick(x, y) {
			return 0
		}
		if app.handleUIStyleClick(x, y) {
			return 0
		}
		if app.handleIntensitySlider(x, y) {
			return 0
		}
		if app.handleModePreviewClick(x, y) {
			return 0
		}
		if app.handleVolumeSlider(x, y) {
			return 0
		}
		if app.handlePlaylistClick(x, y) {
			return 0
		}
		if app.handleButtonClick(x, y) {
			return 0
		}
		if app.seekFromPoint(x, y) {
			app.dragging = true
			procSetCapture.Call(hwnd)
		}
		return 0
	case wmMouseMove:
		app.noteMouseActivity()
		if app.dragging && wParam&mkLButton != 0 {
			x, y := mousePoint(lParam)
			app.seekFromPoint(x, y)
		}
		return 0
	case wmLButtonUp:
		if app.dragging {
			x, y := mousePoint(lParam)
			app.seekFromPoint(x, y)
			app.dragging = false
			procReleaseCapture.Call()
		}
		return 0
	case wmLButtonDblClk:
		x, y := mousePoint(lParam)
		var bounds rect
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
		width := bounds.right - bounds.left
		height := bounds.bottom - bounds.top
		if app.handleVisualizerDoubleClick(x, y, width, height) {
			return 0
		}
		return 0
	case wmPaint:
		draw(hwnd)
		return 0
	case wmDestroy:
		app.saveOverlayBounds()
		app.saveWindowBounds()
		app.removeTray()
		app.stopDesktopInput()
		app.closeOverlay()
		app.stopRemoteControl()
		closeTrack()
		procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
		return ret
	}
}

func (s *appState) openPath(path string) {
	info, err := os.Stat(path)
	if err != nil {
		s.status = "Could not open file: " + err.Error()
		invalidate()
		return
	}
	if info.IsDir() {
		s.openPlaylist(collectAudioFiles(path), 0)
		return
	}
	if !isAudioFile(path) {
		s.status = "Unsupported file. Try WAV, MP3, WMA, MID, AIFF, AU, or SND."
		invalidate()
		return
	}

	playlist := collectAudioFiles(filepath.Dir(path))
	index := indexOfPath(playlist, path)
	if index < 0 {
		playlist = []track{makeTrack(path, s.favorites)}
		index = 0
	}
	s.openPlaylist(playlist, index)
}

func (s *appState) openPlaylist(playlist []track, index int) {
	if len(playlist) == 0 {
		s.status = "No supported audio files found."
		invalidate()
		return
	}
	if index < 0 || index >= len(playlist) {
		index = 0
	}
	s.playlist = playlist
	s.currentIndex = index
	s.libraryRoot = filepath.Dir(playlist[index].Path)
	s.loadCurrentTrack(true)
}

func (s *appState) loadCurrentFolder() {
	if s.filePath == "" {
		s.status = "Open a song first, then press D to load its folder as a playlist."
		invalidate()
		return
	}
	s.openPlaylist(collectAudioFiles(filepath.Dir(s.filePath)), indexOfPath(collectAudioFiles(filepath.Dir(s.filePath)), s.filePath))
}

func (s *appState) loadCurrentTrack(play bool) {
	if s.currentIndex < 0 || s.currentIndex >= len(s.playlist) {
		return
	}
	path := s.playlist[s.currentIndex].Path
	if err := openTrack(path); err != nil {
		s.status = "Windows could not open this audio file: " + err.Error()
		s.frames = nil
		s.playing = false
		closeTrack()
		invalidate()
		return
	}

	duration, err := trackDuration()
	if err != nil || duration <= 0 {
		duration = time.Minute
	}

	s.frames = nil
	s.currentBars = nil
	s.pcmCache = s.loadPCMForPath(path)
	s.ambientSeed = float64(hashPath(path)%1000) / 100
	s.loadTrackMedia(path)
	if frames := s.takePreloadedFrames(path); len(frames) > 0 {
		s.frames = frames
	} else if frames, dur := s.buildFramesForPath(path); len(frames) > 0 {
		s.frames = frames
		if dur > 0 {
			duration = dur
		}
	}

	s.filePath = path
	s.duration = duration
	s.playbackOffset = 0
	s.dragPosition = 0
	s.paused = !play
	s.bumpPlayCount(path)
	s.addRecent(path)
	s.applyVolume()
	s.applyEqualizer()
	s.saveSettings()
	if play {
		s.restart()
	} else {
		s.updateStatus()
		invalidate()
	}
}

func (s *appState) restart() {
	if s.filePath == "" {
		return
	}
	s.seekTo(0, true)
}

func (s *appState) togglePlay() {
	if s.filePath == "" {
		return
	}
	if s.playing {
		pos := s.position()
		if err := pauseTrack(); err != nil {
			s.status = "Pause failed: " + err.Error()
			invalidate()
			return
		}
		s.playbackOffset = pos
		s.playing = false
		s.paused = true
		s.updateStatus()
		return
	}
	if s.position() >= s.duration {
		s.playbackOffset = 0
	}
	s.seekTo(s.playbackOffset, true)
}

func (s *appState) nextTrack() {
	if len(s.playlist) == 0 {
		return
	}
	if s.shuffle && len(s.playlist) > 1 {
		s.currentIndex = s.pickShuffleIndex()
	} else {
		s.currentIndex++
		if s.currentIndex >= len(s.playlist) {
			if s.repeat == repeatAll {
				s.currentIndex = 0
			} else {
				s.currentIndex = len(s.playlist) - 1
				s.playing = false
				s.paused = false
				s.updateStatus()
				invalidate()
				return
			}
		}
	}
	s.loadCurrentTrack(true)
}

func (s *appState) previousTrack() {
	if len(s.playlist) == 0 {
		return
	}
	if s.position() > 3*time.Second {
		s.restart()
		return
	}
	s.currentIndex--
	if s.currentIndex < 0 {
		if s.repeat == repeatAll {
			s.currentIndex = len(s.playlist) - 1
		} else {
			s.currentIndex = 0
		}
	}
	s.loadCurrentTrack(true)
}

func (s *appState) handleTrackEnded() {
	if s.repeat == repeatOne {
		s.restart()
		return
	}
	s.nextTrack()
}

func (s *appState) toggleShuffle() {
	s.shuffle = !s.shuffle
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) cycleRepeat() {
	s.repeat = (s.repeat + 1) % 3
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) toggleFavorite() {
	if s.filePath == "" {
		return
	}
	clean := filepath.Clean(s.filePath)
	s.favorites[clean] = !s.favorites[clean]
	if s.currentIndex >= 0 && s.currentIndex < len(s.playlist) {
		s.playlist[s.currentIndex].Favorite = s.favorites[clean]
	}
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) exportSnapshot() {
	bars := s.targetBars()
	if len(bars) == 0 {
		s.status = "No visualizer snapshot to export yet."
		invalidate()
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, "Desktop")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		dir = home
	}
	name := fmt.Sprintf("music-visualizer-%s.ppm", time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name)
	if err := writeSnapshot(path, bars, s.mode, s.theme); err != nil {
		s.status = "Snapshot failed: " + err.Error()
	} else {
		s.status = "Saved snapshot: " + path
	}
	invalidate()
}

func (s *appState) changeVolume(delta int) {
	s.volume = clampInt(s.volume+delta, 0, 1000)
	s.muted = false
	s.applyVolume()
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) toggleMute() {
	s.muted = !s.muted
	s.applyVolume()
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) toggleDesktopInput() {
	s.cycleInputSource()
}

func (s *appState) stopDesktopInput() {
	if s.desktop != nil {
		s.desktop.stopCapture()
	}
	s.desktopMode = false
	s.saveSettings()
	invalidate()
}

func (s *appState) applyVolume() {
	volume := s.volume
	if s.muted {
		volume = 0
	}
	_ = mciSend(fmt.Sprintf("setaudio musicvisualizer_track volume to %d", volume))
}

func (s *appState) cycleSleepTimer() {
	switch s.sleepMinutes {
	case 0:
		s.sleepMinutes = 15
	case 15:
		s.sleepMinutes = 30
	case 30:
		s.sleepMinutes = 60
	default:
		s.sleepMinutes = 0
	}
	s.sleepStartedAt = time.Now()
	s.updateStatus()
	invalidate()
}

func (s *appState) tickSleepTimer() {
	if s.sleepMinutes == 0 || s.sleepStartedAt.IsZero() {
		return
	}
	if time.Since(s.sleepStartedAt) >= time.Duration(s.sleepMinutes)*time.Minute {
		if s.playing {
			s.togglePlay()
		}
		s.sleepMinutes = 0
		s.status = "Sleep timer paused playback."
	}
}

func (s *appState) toggleMini() {
	s.mini = !s.mini
	s.updateStatus()
	invalidate()
}

func (s *appState) toggleFullscreen() {
	s.fullscreen = !s.fullscreen
	if s.fullscreen {
		procShowWindow.Call(s.hwnd, swShowMaximized)
	} else {
		procShowWindow.Call(s.hwnd, swShowNormal)
	}
	s.updateStatus()
	invalidate()
}

func (s *appState) seekBy(delta time.Duration) {
	if s.filePath == "" {
		return
	}
	s.seekTo(s.position()+delta, s.playing)
}

func (s *appState) seekTo(pos time.Duration, play bool) {
	if s.filePath == "" {
		return
	}
	pos = clampDuration(pos, 0, s.duration)
	s.playbackOffset = pos
	s.dragPosition = pos
	if play {
		if err := playTrackFrom(pos); err != nil {
			s.status = "Play failed: " + err.Error()
			s.playing = false
			s.paused = false
			invalidate()
			return
		}
		s.startedAt = time.Now()
		s.playing = true
		s.paused = false
		s.applyPlaybackSpeed()
	} else if !s.paused {
		_ = stopTrack()
	}
	s.updateStatus()
}

func (s *appState) cycleMode() {
	next := (s.mode + 1) % modeCount
	s.setMode(next)
	s.flashStatus("Visualizer: "+modeName(s.mode), 2*time.Second)
	s.saveSettings()
	invalidate()
}

func (s *appState) nextTheme() {
	maxTheme := themeOcean + 1
	if len(s.customThemes) > 0 {
		maxTheme = themeCustomBase + theme(len(s.customThemes))
	}
	s.theme = (s.theme + 1) % maxTheme
	s.updateStatus()
	s.saveSettings()
	invalidate()
}

func (s *appState) cycleEqualizer() {
	s.focusNextEQBand()
}

func (s *appState) seekStep() time.Duration {
	if s.seekLarge {
		return 30 * time.Second
	}
	return 10 * time.Second
}

func (s *appState) position() time.Duration {
	if s.dragging {
		return s.dragPosition
	}
	if !s.playing {
		return clampDuration(s.playbackOffset, 0, s.duration)
	}
	return clampDuration(s.playbackOffset+time.Since(s.startedAt), 0, s.duration)
}

func (s *appState) progress() float64 {
	if s.duration <= 0 {
		return 0
	}
	return clamp(float64(s.position())/float64(s.duration), 0, 1)
}

func (s *appState) updateStatus() {
	if s.filePath == "" {
		lib := ""
		if s.libraryIndex != nil {
			lib = libraryStatusText(s.libraryScanning, len(s.libraryIndex.Entries)) + " | "
		}
		s.status = lib + fmt.Sprintf("Open or drag audio. Input:%s | Mode:%s | Theme:%s | Volume:%d%%", s.inputName(), modeName(s.mode), themeName(s.theme), s.volume/10)
		return
	}
	state := "Playing"
	if s.paused {
		state = "Paused"
	} else if !s.playing {
		state = "Ready"
	}
	fav := ""
	if s.favorites[filepath.Clean(s.filePath)] {
		fav = " *Favorite"
	}
	s.status = fmt.Sprintf("%s %s%s  %s/%s  Input:%s  Mode:%s  Theme:%s  Vol:%d%%  Shuffle:%t  Repeat:%s  Sleep:%s  Mood:%s",
		state,
		filepath.Base(s.filePath),
		fav,
		durationText(s.position()),
		durationText(s.duration),
		s.inputName(),
		modeName(s.mode),
		themeName(s.theme),
		s.volume/10,
		s.shuffle,
		repeatName(s.repeat),
		sleepText(s.sleepMinutes),
		s.mood,
	)
	if dj := djStatusText(s.djMode, s.djCrossfader, s.djDeckBPath); dj != "" {
		s.status += "  " + dj
	}
}

func (s *appState) inputName() string {
	return s.inputSourceName()
}

func (s *appState) targetBars() []float64 {
	bars := s.rawBars()
	if s.djMode {
		bars = s.mixDJBars(bars)
	}
	return s.applyVisualEQ(bars)
}

func (s *appState) rawBars() []float64 {
	if s.desktopMode && s.desktop != nil && s.desktop.isActive() {
		if snap := s.desktop.snapshot(); len(snap) > 0 {
			return snap
		}
	}
	if s.liveFFT && !s.desktopMode && s.pcmCache != nil && s.playing {
		if bars := s.liveFFTBars(); len(bars) > 0 {
			return bars
		}
	}
	if len(s.frames) == 0 {
		return s.ambientBars()
	}
	pos := s.position()
	if s.playing && pos >= s.duration {
		s.playing = false
		s.playbackOffset = s.duration
		_ = stopTrack()
		s.updateStatus()
		return s.frames[len(s.frames)-1].Bars
	}
	index := int(pos.Seconds() * fps)
	if index < 0 {
		index = 0
	}
	if index >= len(s.frames) {
		index = len(s.frames) - 1
	}
	return s.frames[index].Bars
}

func (s *appState) ambientBars() []float64 {
	bars := make([]float64, barCount)
	pos := s.position().Seconds()
	if !s.playing {
		pos = float64(time.Now().UnixMilli()) / 1000
	}
	pulse := 0.5 + math.Sin(pos*2.4+s.ambientSeed)*0.5
	for i := range bars {
		x := float64(i)
		low := math.Sin(pos*3.1+s.ambientSeed+x*0.17)*0.5 + 0.5
		mid := math.Sin(pos*5.7+s.ambientSeed*0.6+x*0.31)*0.5 + 0.5
		high := math.Sin(pos*8.3+s.ambientSeed*1.3+x*0.73)*0.5 + 0.5
		shape := math.Sin(float64(i) / float64(barCount) * math.Pi)
		bars[i] = clamp((low*0.45+mid*0.35+high*0.20)*(0.35+shape*0.65)+pulse*0.18, 0.04, 1)
	}
	return bars
}

func openWAVDialog(hwnd uintptr) (string, bool) {
	var fileBuffer [4096]uint16
	filter := utf16WithNULs("Audio files\x00*.wav;*.mp3;*.flac;*.ogg;*.oga;*.wma;*.mid;*.midi;*.aiff;*.aif;*.au;*.snd\x00WAV files (*.wav)\x00*.wav\x00MP3 files (*.mp3)\x00*.mp3\x00FLAC files (*.flac)\x00*.flac\x00OGG files (*.ogg)\x00*.ogg\x00All files (*.*)\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("Open an audio file")

	ofn := openFileName{
		structSize: uint32(unsafe.Sizeof(openFileName{})),
		hwndOwner:  hwnd,
		filter:     &filter[0],
		file:       &fileBuffer[0],
		maxFile:    uint32(len(fileBuffer)),
		title:      title,
		flags:      ofnExplorer | ofnPathMustExist | ofnFileMustExist,
	}

	ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", false
	}
	return syscall.UTF16ToString(fileBuffer[:]), true
}

func draw(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	var bounds rect
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))

	width := bounds.right - bounds.left
	height := bounds.bottom - bounds.top
	if width <= 0 || height <= 0 {
		return
	}

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return
	}
	defer procDeleteDC.Call(memDC)

	bitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(width), uintptr(height))
	if bitmap == 0 {
		return
	}
	defer procDeleteObject.Call(bitmap)

	oldBitmap, _, _ := procSelectObject.Call(memDC, bitmap)
	defer procSelectObject.Call(memDC, oldBitmap)

	drawFrame(memDC, width, height)
	procBitBlt.Call(hdc, 0, 0, uintptr(width), uintptr(height), memDC, 0, 0, srccopy)
}

func drawFrame(hdc uintptr, width, height int32) {
	palette := currentPalette()
	if app.moodReactive {
		palette = app.moodAdjustedPalette(palette)
	}
	if app.ultra != nil && app.ultra.artTintOn && app.artGrid != nil && app.uiStyle != uiStyleLight {
		palette = app.ultra.artTint
		if app.moodReactive {
			palette = app.moodAdjustedPalette(palette)
		}
	}
	palette = uiPaletteForStyle(app.uiStyle, palette)
	app.uiLayout = layoutFor(app.uiStyle, width, height)
	chrome := app.uiLayout

	fill(hdc, rect{left: 0, top: 0, right: width, bottom: height}, palette.background)

	procSetBkMode.Call(hdc, transparent)
	procSetTextColor.Call(hdc, palette.text)
	textOut(hdc, chrome.marginL, chrome.titleY, appTitle+" Ultra")
	if app.ultra != nil && app.ultra.showcase && !app.cinemaUIVisible() {
		procSetTextColor.Call(hdc, palette.dim)
		textOut(hdc, chrome.marginL, height-36, "Cinema mode — move mouse for controls, Esc to exit, F9 toggle")
	} else if !app.mini && !app.visualOnly {
		procSetTextColor.Call(hdc, palette.dim)
		hasMeta := app.meta.Title != ""
		if hasMeta {
			metaLine := fmt.Sprintf("%s — %s", app.meta.Title, app.meta.Artist)
			if line := app.currentLyric(); line != "" {
				metaLine += " | " + line
			}
			textOut(hdc, chrome.marginL, chrome.metaY, metaLine)
			if chrome.metaY != chrome.statusY {
				textOut(hdc, chrome.marginL, chrome.statusY, app.status)
			}
		} else {
			textOut(hdc, chrome.marginL, chrome.statusY, app.status)
		}
		if chrome.showHints {
			textOut(hdc, chrome.marginL, height-30, "F8 UI style | F9 cinema | 1 OBS | 3 settings | V viz")
		}
		drawButtons(hdc, width, height, palette, chrome)
		if chrome.showVolume {
			app.drawVolumeSlider(hdc, width, palette, chrome)
		}
		drawPlaylist(hdc, width, height, palette, chrome)
		app.drawSettingsPanel(hdc, width, height, palette, chrome)
	}
	if line := app.karaokeLine(); line != "" && !app.visualOnly {
		procSetTextColor.Call(hdc, palette.accent2)
		textOut(hdc, chrome.marginL, height-52, line)
	}
	if !app.visualOnly {
		drawProgress(hdc, width, palette, chrome)
	}

	bars := app.targetBars()
	app.updateAnalysis(bars)
	if app.onset != nil {
		app.onset.Update(bars, app.bpm, 1.0/float64(fps))
	}
	if app.mode == modeSpectrogram {
		app.pushSpecHistory(bars)
	}
	if app.perf != nil {
		app.perf.tickFrame()
	}
	app.publishBarsStream(bars)
	beat := app.beatMultiplier(bars)
	if app.onset != nil {
		beat = math.Max(beat, app.onset.CombinedBeat())
	}
	intensity := app.visualIntensity
	if app.ambientMode {
		intensity *= 0.65
	}
	if app.ultra != nil {
		bars = app.ultra.processBars(bars, intensity, beat)
		bars = app.applyBPMPulse(bars)
		app.currentBars = bars
	} else {
		for i := range bars {
			bars[i] = math.Min(1, bars[i]*beat*intensity)
		}
		if len(app.currentBars) != len(bars) {
			app.currentBars = make([]float64, len(bars))
		}
		smooth := 0.35
		if app.modeBlend < 1 {
			smooth = 0.18
		}
		if app.ambientMode {
			smooth = 0.12
		}
		for i, target := range bars {
			app.currentBars[i] += (target - app.currentBars[i]) * smooth
		}
	}

	top := chrome.vizTop
	bottom := chrome.vizBottom
	if app.mini || app.visualOnly {
		top = 28
		bottom = height - 28
	}
	if bottom <= top {
		return
	}

	visualBounds := chrome.visualBounds(width, height)
	if app.artGrid != nil && !app.visualOnly {
		drawAlbumArtBackground(hdc, visualBounds, app.artGrid)
	}
	if app.ultra != nil {
		if app.modeBlend < 1 && app.prevMode != app.mode {
			drawWithEffectsBlend(hdc, visualBounds, app.currentBars, app.ultra.peakBars(), app.ultra.trailFrames(), app.prevMode, app.mode, app.modeBlend)
		} else {
			drawWithEffects(hdc, visualBounds, app.currentBars, app.ultra.peakBars(), app.ultra.trailFrames(), app.mode)
		}
	} else if app.modeBlend < 1 && app.prevMode != app.mode {
		drawVisualizationBlend(hdc, visualBounds, app.currentBars, app.prevMode, app.mode, app.modeBlend)
	} else {
		drawVisualization(hdc, visualBounds, app.currentBars, app.mode)
	}
}

func drawVisualization(hdc uintptr, bounds rect, bars []float64, mode visualMode) {
	if isShaderMode(mode) {
		if drawShaderMode(hdc, bounds, bars, mode) {
			return
		}
	}
	switch mode {
	case modeMirror:
		drawMirrorBars(hdc, bounds, bars)
	case modeBlocks:
		drawBlockStack(hdc, bounds, bars)
	case modeWave:
		drawWaveLine(hdc, bounds, bars)
	case modeHalo:
		drawHalo(hdc, bounds, bars)
	case modeSpectrum:
		drawSpectrum(hdc, bounds, bars)
	case modeFire:
		drawFire(hdc, bounds, bars)
	case modeParticles:
		drawParticles(hdc, bounds, bars)
	case modeTunnel:
		drawTunnel(hdc, bounds, bars)
	case modePlasma:
		drawPlasma(hdc, bounds, bars)
	case modeAurora:
		drawAurora(hdc, bounds, bars)
	case modeMandala:
		drawMandala(hdc, bounds, bars)
	case modeStarfield:
		drawStarfield(hdc, bounds, bars)
	case modeKaleidoscope:
		drawKaleidoscope(hdc, bounds, bars)
	case modeNeonCity:
		drawNeonCity(hdc, bounds, bars)
	case modeSupernova:
		drawSupernova(hdc, bounds, bars)
	case modeLiquid:
		drawLiquid(hdc, bounds, bars)
	case modeOrbit:
		drawOrbit(hdc, bounds, bars)
	case modeWaveform3D:
		drawWaveform3D(hdc, bounds, bars)
	case modeSpectrogram:
		drawSpectrogram(hdc, bounds, bars)
	case modeOscilloscope:
		drawOscilloscope(hdc, bounds, bars)
	case modeLissajous:
		drawLissajous(hdc, bounds, bars)
	default:
		drawClassicBars(hdc, bounds, bars)
	}
}

func drawClassicBars(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	margin := int32(0)
	gap := int32(5)
	width := bounds.right - bounds.left
	usableHeight := bounds.bottom - bounds.top
	barWidth := (width - margin*2 - gap*int32(len(bars)-1)) / int32(len(bars))
	if barWidth < 2 {
		barWidth = 2
		gap = 2
	}
	for i, target := range bars {
		value := math.Pow(clamp(target, 0, 1), 0.75)
		barHeight := int32(value * float64(usableHeight))
		left := bounds.left + margin + int32(i)*(barWidth+gap)
		right := left + barWidth
		color := bandGradientColor(i, len(bars), bars)
		fill(hdc, rect{left: left, top: bounds.bottom - barHeight, right: right, bottom: bounds.bottom}, color)
	}
}

func drawMirrorBars(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	gap := int32(5)
	width := bounds.right - bounds.left
	middle := bounds.top + (bounds.bottom-bounds.top)/2
	halfHeight := (bounds.bottom - bounds.top) / 2
	barWidth := (width - gap*int32(len(bars)-1)) / int32(len(bars))
	if barWidth < 2 {
		barWidth = 2
		gap = 2
	}
	for i, target := range bars {
		value := math.Pow(clamp(target, 0, 1), 0.8)
		barHeight := int32(value * float64(halfHeight))
		left := bounds.left + int32(i)*(barWidth+gap)
		right := left + barWidth
		color := gradientColor(i, len(bars), 230)
		fill(hdc, rect{left: left, top: middle - barHeight, right: right, bottom: middle}, color)
		fill(hdc, rect{left: left, top: middle + 2, right: right, bottom: middle + barHeight + 2}, dimColor(color, 0.55))
	}
	line(hdc, bounds.left, middle, bounds.right, middle, rgb(34, 44, 70), 1)
}

func drawBlockStack(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	const blocks = 12
	gap := int32(4)
	width := bounds.right - bounds.left
	barWidth := (width - gap*int32(len(bars)-1)) / int32(len(bars))
	blockGap := int32(4)
	blockHeight := (bounds.bottom - bounds.top - blockGap*(blocks-1)) / blocks
	if blockHeight < 3 {
		blockHeight = 3
		blockGap = 2
	}
	for i, target := range bars {
		filled := int(math.Ceil(clamp(target, 0, 1) * blocks))
		left := bounds.left + int32(i)*(barWidth+gap)
		for block := 0; block < blocks; block++ {
			top := bounds.bottom - int32(block+1)*(blockHeight+blockGap)
			if block < filled {
				fill(hdc, rect{left: left, top: top, right: left + barWidth, bottom: top + blockHeight}, gradientColor(i+block, len(bars)+blocks, 255))
			} else if block%3 == 0 {
				fill(hdc, rect{left: left, top: top, right: left + barWidth, bottom: top + blockHeight}, rgb(20, 25, 43))
			}
		}
	}
}

func drawWaveLine(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	middle := bounds.top + (bounds.bottom-bounds.top)/2
	width := bounds.right - bounds.left
	line(hdc, bounds.left, middle, bounds.right, middle, rgb(30, 40, 64), 1)
	for layer := 0; layer < 3; layer++ {
		color := []uintptr{rgb(80, 220, 255), rgb(185, 90, 255), rgb(255, 90, 170)}[layer]
		penWidth := int32(4 - layer)
		pen, _, _ := procCreatePen.Call(0, uintptr(penWidth), color)
		if pen == 0 {
			continue
		}
		old, _, _ := procSelectObject.Call(hdc, pen)
		for i, target := range bars {
			x := bounds.left + int32(i)*width/int32(max(1, len(bars)-1))
			phase := math.Sin(float64(i+layer*5) * 0.45)
			y := middle - int32((target*0.75+phase*0.08)*float64(bounds.bottom-bounds.top)/2)
			if i == 0 {
				procMoveToEx.Call(hdc, uintptr(x), uintptr(y), 0)
			} else {
				procLineTo.Call(hdc, uintptr(x), uintptr(y))
			}
		}
		procSelectObject.Call(hdc, old)
		procDeleteObject.Call(pen)
	}
}

func drawHalo(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	maxRadius := min(bounds.right-bounds.left, bounds.bottom-bounds.top) / 2
	for ring := 3; ring >= 1; ring-- {
		radius := maxRadius * int32(ring) / 4
		ellipse(hdc, cx-radius, cy-radius, cx+radius, cy+radius, rgb(22+byte(ring*12), 28+byte(ring*12), 48+byte(ring*14)))
	}
	for i, target := range bars {
		angle := float64(i)/float64(len(bars))*math.Pi*2 - math.Pi/2
		inner := float64(maxRadius) * 0.42
		outer := inner + math.Pow(clamp(target, 0, 1), 0.65)*float64(maxRadius)*0.56
		x1 := cx + int32(math.Cos(angle)*inner)
		y1 := cy + int32(math.Sin(angle)*inner)
		x2 := cx + int32(math.Cos(angle)*outer)
		y2 := cy + int32(math.Sin(angle)*outer)
		line(hdc, x1, y1, x2, y2, gradientColor(i, len(bars), 255), 3)
	}
}

func drawSpectrum(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	bands := []uintptr{rgb(66, 220, 255), rgb(120, 255, 120), rgb(255, 220, 70), rgb(255, 88, 88)}
	gap := int32(4)
	width := bounds.right - bounds.left
	barWidth := (width - gap*int32(len(bars)-1)) / int32(len(bars))
	for i, target := range bars {
		value := math.Pow(clamp(target, 0, 1), 0.7)
		barHeight := int32(value * float64(bounds.bottom-bounds.top))
		left := bounds.left + int32(i)*(barWidth+gap)
		band := i * len(bands) / len(bars)
		fill(hdc, rect{left: left, top: bounds.bottom - barHeight, right: left + barWidth, bottom: bounds.bottom}, bands[band])
	}
}

func drawFire(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	const layers = 9
	width := bounds.right - bounds.left
	cellW := max(int32(3), width/int32(len(bars)))
	for i, target := range bars {
		flame := int(clamp(target, 0, 1) * layers)
		for layer := 0; layer < layers; layer++ {
			left := bounds.left + int32(i)*cellW
			bottom := bounds.bottom - int32(layer)*(bounds.bottom-bounds.top)/layers
			top := bottom - (bounds.bottom-bounds.top)/layers + 2
			color := rgb(byte(80+layer*19), byte(20+layer*18), byte(max(0, 80-layer*10)))
			if layer < flame {
				fill(hdc, rect{left: left, top: top, right: left + cellW - 1, bottom: bottom}, color)
			}
		}
	}
}

func drawParticles(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	now := float64(time.Now().UnixMilli()) / 700
	for i, target := range bars {
		angle := float64(i)/float64(len(bars))*math.Pi*2 + now*0.18
		radius := (0.15 + target*0.82) * float64(min(bounds.right-bounds.left, bounds.bottom-bounds.top)) / 2
		x := cx + int32(math.Cos(angle)*radius)
		y := cy + int32(math.Sin(angle)*radius)
		size := int32(3 + target*10)
		ellipse(hdc, x-size, y-size, x+size, y+size, gradientColor(i, len(bars), 255))
	}
}

func drawTunnel(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cx := bounds.left + (bounds.right-bounds.left)/2
	cy := bounds.top + (bounds.bottom-bounds.top)/2
	maxR := min(bounds.right-bounds.left, bounds.bottom-bounds.top) / 2
	now := float64(time.Now().UnixMilli()) / 800
	for ring := 9; ring >= 1; ring-- {
		idx := ring * len(bars) / 10
		energy := bars[min(idx, len(bars)-1)]
		radius := int32(float64(maxR) * (float64(ring)/10 + energy*0.06))
		offset := int32(math.Sin(now+float64(ring)) * energy * 18)
		line(hdc, cx-radius+offset, cy-radius, cx+radius, cy-radius+offset, gradientColor(ring, 10, 255), 2)
		line(hdc, cx+radius, cy-radius+offset, cx+radius-offset, cy+radius, gradientColor(ring, 10, 255), 2)
		line(hdc, cx+radius-offset, cy+radius, cx-radius, cy+radius-offset, gradientColor(ring, 10, 255), 2)
		line(hdc, cx-radius, cy+radius-offset, cx-radius+offset, cy-radius, gradientColor(ring, 10, 255), 2)
	}
}

func drawPlasma(hdc uintptr, bounds rect, bars []float64) {
	if len(bars) == 0 {
		return
	}
	cols := 18
	rows := 10
	cellW := max(int32(3), (bounds.right-bounds.left)/int32(cols))
	cellH := max(int32(3), (bounds.bottom-bounds.top)/int32(rows))
	now := float64(time.Now().UnixMilli()) / 520
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			idx := (x + y*cols) % len(bars)
			energy := bars[idx]
			wave := math.Sin(float64(x)*0.9+now) + math.Cos(float64(y)*1.2-now*0.8)
			level := clamp((wave+2)/4*0.55+energy*0.45, 0, 1)
			left := bounds.left + int32(x)*cellW
			top := bounds.top + int32(y)*cellH
			fill(hdc, rect{left: left, top: top, right: left + cellW - 2, bottom: top + cellH - 2}, rgb(byte(45+level*160), byte(40+level*80), byte(110+level*140)))
		}
	}
}

func drawProgress(hdc uintptr, width int32, palette palette, chrome uiChromeLayout) {
	bar := chrome.progressBar(width)
	fillPanel(hdc, bar, palette, chrome)
	progress := app.progress()
	fill(hdc, rect{left: bar.left, top: bar.top, right: bar.left + int32(progress*float64(bar.right-bar.left)), bottom: bar.bottom}, palette.accent)
	knobX := bar.left + int32(progress*float64(bar.right-bar.left))
	fill(hdc, rect{left: knobX - 4, top: bar.top - 5, right: knobX + 4, bottom: bar.bottom + 5}, palette.text)
	if chrome.showHints && !app.mini {
		procSetTextColor.Call(hdc, palette.dim)
		textOut(hdc, bar.left, bar.bottom+14, "Click/drag to seek. Drop audio files anywhere.")
	}
}

func drawButtons(hdc uintptr, width, height int32, palette palette, chrome uiChromeLayout) {
	labels := []button{
		{label: "Open", action: "open"},
		{label: "Prev", action: "prev"},
		{label: playLabel(), action: "play"},
		{label: "Next", action: "next"},
		{label: "Viz", action: "viz"},
		{label: inputLabel(), action: "input"},
		{label: "Theme", action: "theme"},
		{label: "Lib", action: "library"},
		{label: "Preset", action: "preset"},
		{label: "OBS", action: "overlay"},
		{label: "DJ", action: "dj"},
		{label: "Set", action: "settings"},
		{label: "Fav", action: "favorite"},
		{label: "Mute", action: "mute"},
	}
	app.buttons = app.buttons[:0]
	left := chrome.marginL
	top := chrome.buttonTop
	for i := range labels {
		w := chrome.buttonW
		if left+w > width-chrome.marginR {
			if app.uiStyle == uiStyleCompact {
				left = chrome.marginL
				top += chrome.buttonH + 4
			} else {
				break
			}
		}
		b := labels[i]
		b.bounds = rect{left: left, top: top, right: left + w, bottom: top + chrome.buttonH}
		app.buttons = append(app.buttons, b)
		fillPanel(hdc, b.bounds, palette, chrome)
		line(hdc, b.bounds.left, b.bounds.bottom, b.bounds.right, b.bounds.bottom, palette.accent, 2)
		procSetTextColor.Call(hdc, palette.text)
		textOut(hdc, b.bounds.left+8, b.bounds.top+5, b.label)
		left += w + chrome.buttonGap
	}
}

func drawPlaylist(hdc uintptr, width, height int32, palette palette, chrome uiChromeLayout) {
	list := app.filteredPlaylist()
	groups := app.libraryGroupsForPanel()
	entries := app.libraryEntriesForPanel()
	count := len(list)
	if app.panel >= viewLibrary {
		if len(groups) > 0 && app.panelGroupKey == "" {
			count = len(groups)
		} else if len(entries) > 0 {
			count = len(entries)
		}
	}
	if count == 0 && app.panel == viewQueue {
		if !chrome.showPlaylist {
			return
		}
		left := chrome.playlistLeft
		if left < chrome.marginL {
			return
		}
		top := chrome.playlistTop
		panel := rect{left: left, top: top, right: left + chrome.playlistW, bottom: chrome.vizBottom}
		fillPanel(hdc, panel, palette, chrome)
		procSetTextColor.Call(hdc, palette.text)
		textOut(hdc, left+12, top+10, "Queue (0)")
		procSetTextColor.Call(hdc, palette.dim)
		textOut(hdc, left+12, top+38, "Drop audio files here")
		textOut(hdc, left+12, top+58, "or press O to open")
		return
	}
	if !chrome.showPlaylist && app.panel == viewQueue {
		return
	}
	if !chrome.showPlaylist && count == 0 {
		return
	}
	left := chrome.playlistLeft
	if left < chrome.marginL {
		return
	}
	top := chrome.playlistTop
	panel := rect{left: left, top: top, right: left + chrome.playlistW, bottom: chrome.vizBottom}
	fillPanel(hdc, panel, palette, chrome)
	procSetTextColor.Call(hdc, palette.text)
	textOut(hdc, left+12, top+10, fmt.Sprintf("%s (%d)", app.panelTitle(), count))
	app.playlistRowBounds = app.playlistRowBounds[:0]
	procSetTextColor.Call(hdc, palette.dim)
	if len(groups) > 0 && app.panelGroupKey == "" {
		for row := 0; row < 8 && row < len(groups); row++ {
			y := top + 38 + int32(row*24)
			rowRect := rect{left: left + 8, top: y - 4, right: left + chrome.playlistW - 12, bottom: y + 18}
			app.playlistRowBounds = append(app.playlistRowBounds, rowRect)
			textOut(hdc, left+12, y, truncate(groups[row], 34))
		}
		return
	}
	if len(entries) > 0 && app.panel >= viewLibrary {
		for row := 0; row < 8 && row < len(entries); row++ {
			e := entries[row]
			y := top + 38 + int32(row*24)
			rowRect := rect{left: left + 8, top: y - 4, right: left + chrome.playlistW - 12, bottom: y + 18}
			app.playlistRowBounds = append(app.playlistRowBounds, rowRect)
			label := e.Title
			if e.PlayCount > 0 {
				label = fmt.Sprintf("%s (%d)", truncate(e.Title, 26), e.PlayCount)
			}
			textOut(hdc, left+12, y, truncate(label, 34))
		}
		return
	}
	for row := 0; row < 8 && row < len(list); row++ {
		t := list[row]
		y := top + 38 + int32(row*24)
		rowRect := rect{left: left + 8, top: y - 4, right: left + chrome.playlistW - 12, bottom: y + 18}
		app.playlistRowBounds = append(app.playlistRowBounds, rowRect)
		fav := ""
		if t.Favorite {
			fav = "* "
		}
		textOut(hdc, left+12, y, truncate(fav+t.Title, 34))
	}
}

func (s *appState) handleButtonClick(x, y int32) bool {
	for _, b := range s.buttons {
		if pointInRect(x, y, b.bounds) {
			switch b.action {
			case "open":
				if path, ok := openWAVDialog(s.hwnd); ok {
					s.openPath(path)
				}
			case "prev":
				s.previousTrack()
			case "play":
				s.togglePlay()
			case "next":
				s.nextTrack()
			case "viz":
				s.cycleMode()
			case "input":
				s.toggleDesktopInput()
			case "theme":
				s.nextTheme()
			case "library":
				s.scanLibraryAsync()
				s.showLibraryView()
			case "preset":
				s.cyclePreset()
			case "overlay":
				s.toggleOverlay()
			case "dj":
				s.toggleDJMode()
			case "settings":
				s.toggleSettingsPanel()
			case "shuffle":
				s.toggleShuffle()
			case "repeat":
				s.cycleRepeat()
			case "favorite":
				s.toggleFavorite()
			case "mute":
				s.toggleMute()
			case "speed":
				s.cyclePlaybackSpeed()
			case "recent":
				s.showRecentView()
			case "favs":
				s.showFavoritesView()
			}
			return true
		}
	}
	return false
}

func (s *appState) handleDrop(drop uintptr) {
	defer procDragFinish.Call(drop)
	count, _, _ := procDragQueryFileW.Call(drop, ^uintptr(0), 0, 0)
	var tracks []track
	for i := uintptr(0); i < count; i++ {
		var buffer [4096]uint16
		procDragQueryFileW.Call(drop, i, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
		path := syscall.UTF16ToString(buffer[:])
		if path == "" {
			continue
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			tracks = append(tracks, collectAudioFiles(path)...)
		} else if isAudioFile(path) {
			tracks = append(tracks, makeTrack(path, s.favorites))
		}
	}
	if len(tracks) > 0 {
		sortTracks(tracks)
		s.openPlaylist(tracks, 0)
	}
}

func collectAudioFiles(root string) []track {
	var tracks []track
	entries, err := os.ReadDir(root)
	if err != nil {
		return tracks
	}
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			continue
		}
		if isAudioFile(path) {
			tracks = append(tracks, makeTrack(path, app.favorites))
		}
	}
	sortTracks(tracks)
	return tracks
}

func sortTracks(tracks []track) {
	sort.Slice(tracks, func(i, j int) bool {
		return strings.ToLower(tracks[i].Title) < strings.ToLower(tracks[j].Title)
	})
}

func makeTrack(path string, favorites map[string]bool) track {
	clean := filepath.Clean(path)
	title := strings.TrimSuffix(filepath.Base(clean), filepath.Ext(clean))
	return track{Path: clean, Title: title, Favorite: favorites[clean]}
}

func indexOfPath(tracks []track, path string) int {
	clean := filepath.Clean(path)
	for i, track := range tracks {
		if strings.EqualFold(track.Path, clean) {
			return i
		}
	}
	return -1
}

func isAudioFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".wav", ".mp3", ".flac", ".ogg", ".oga", ".wma", ".mid", ".midi", ".aiff", ".aif", ".au", ".snd":
		return true
	default:
		return false
	}
}

func (s *appState) addRecent(path string) {
	clean := filepath.Clean(path)
	next := []string{clean}
	for _, item := range s.recent {
		if !strings.EqualFold(item, clean) && len(next) < 10 {
			next = append(next, item)
		}
	}
	s.recent = next
}

func (s *appState) loadSettings() {
	s.favorites = map[string]bool{}
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return
	}
	var cfg settings
	if json.Unmarshal(data, &cfg) != nil {
		return
	}
	s.mode = cfg.Mode % modeCount
	s.prevMode = s.mode
	s.modeBlend = 1
	s.theme = cfg.Theme % themeCount
	s.volume = clampInt(cfg.Volume, 0, 1000)
	if s.volume == 0 {
		s.volume = 800
	}
	s.muted = cfg.Muted
	s.desktopMode = cfg.DesktopInput
	s.desktopAuto = cfg.DesktopAuto
	if cfg.DesktopSensitivity > 0 {
		s.desktopSensitivity = cfg.DesktopSensitivity
	}
	s.desktopDevice = cfg.DesktopDevice
	s.repeat = cfg.Repeat % 3
	s.shuffle = cfg.Shuffle
	s.recent = cfg.Recent
	s.libraryRoot = cfg.LibraryRoot
	s.seenOnboarding = cfg.SeenOnboarding
	s.windowX = cfg.WindowX
	s.windowY = cfg.WindowY
	s.windowW = cfg.WindowW
	s.windowH = cfg.WindowH
	if cfg.PlaybackSpeed > 0 {
		s.playbackSpeed = cfg.PlaybackSpeed
	}
	s.eqBass = cfg.EqBass
	s.eqMid = cfg.EqMid
	s.eqTreble = cfg.EqTreble
	s.customThemes = cfg.CustomThemes
	s.presetIndex = cfg.PresetIndex
	s.presetsFromSettings(cfg.Presets)
	if cfg.VisualIntensity > 0 {
		s.visualIntensity = cfg.VisualIntensity
	}
	s.playCounts = cfg.PlayCounts
	if s.playCounts == nil {
		s.playCounts = map[string]int{}
	}
	s.lastPlayed = cfg.LastPlayed
	if s.lastPlayed == nil {
		s.lastPlayed = map[string]int64{}
	}
	s.libraryPaths = cfg.LibraryPaths
	s.crossfade = cfg.Crossfade
	s.karaokeMode = cfg.KaraokeMode
	s.ambientMode = cfg.AmbientMode
	s.overlayMode = cfg.OverlayMode
	s.overlayX = cfg.OverlayX
	s.overlayY = cfg.OverlayY
	s.overlayW = cfg.OverlayW
	s.overlayH = cfg.OverlayH
	if cfg.OverlayAlpha >= 80 && cfg.OverlayAlpha <= 255 {
		s.overlayAlpha = byte(cfg.OverlayAlpha)
	}
	if cfg.RemoteControl {
		s.remoteEnabled = true
		s.startRemoteControl()
	} else {
		s.remoteEnabled = false
	}
	s.autoPreset = cfg.AutoPreset
	s.djMode = cfg.DJMode
	s.panelGroup = cfg.PanelGroup
	s.uiStyle = uiStyle(cfg.UIStyle % int(uiStyleCount))
	s.liveFFT = true
	s.moodReactive = true
	if cfg.SettingsVersion >= 2 {
		s.liveFFT = cfg.LiveFFT
		s.moodReactive = cfg.MoodReactive
	}
	s.inputSource = inputSource(cfg.InputSource % 3)
	s.overlayAspect = overlayAspect(cfg.OverlayAspect % 4)
	s.chromaKey = cfg.ChromaKey
	for _, fav := range cfg.Favorites {
		s.favorites[filepath.Clean(fav)] = true
	}
}

func (s *appState) saveSettings() {
	cfg := settings{
		Mode:               s.mode,
		Theme:              s.theme,
		Volume:             s.volume,
		Muted:              s.muted,
		DesktopInput:       s.desktopMode,
		DesktopAuto:        s.desktopAuto,
		DesktopSensitivity: s.desktopSensitivity,
		DesktopDevice:      s.desktopDevice,
		Repeat:             s.repeat,
		Shuffle:            s.shuffle,
		Recent:             s.recent,
		LibraryRoot:        s.libraryRoot,
		SeenOnboarding:     s.seenOnboarding,
		WindowX:            s.windowX,
		WindowY:            s.windowY,
		WindowW:            s.windowW,
		WindowH:            s.windowH,
		PlaybackSpeed:      s.playbackSpeed,
		EqBass:             s.eqBass,
		EqMid:              s.eqMid,
		EqTreble:           s.eqTreble,
		CustomThemes:       s.customThemes,
		Presets:            s.presetsToSettings(),
		PresetIndex:        s.presetIndex,
		PlayCounts:         s.playCounts,
		LastPlayed:         s.lastPlayed,
		LibraryPaths:       s.libraryPaths,
		Crossfade:          s.crossfade,
		KaraokeMode:        s.karaokeMode,
		AmbientMode:        s.ambientMode,
		OverlayMode:        s.overlayMode,
		OverlayX:           s.overlayX,
		OverlayY:           s.overlayY,
		OverlayW:           s.overlayW,
		OverlayH:           s.overlayH,
		OverlayAlpha:       int(s.overlayAlpha),
		RemoteControl:      s.remoteEnabled,
		AutoPreset:         s.autoPreset,
		DJMode:             s.djMode,
		VisualIntensity:    s.visualIntensity,
		UIStyle:            int(s.uiStyle),
		PanelGroup:         s.panelGroup,
		PanelGroupKey:      s.panelGroupKey,
		SettingsVersion:    2,
		LiveFFT:            s.liveFFT,
		MoodReactive:       s.moodReactive,
		InputSource:        int(s.inputSource),
		OverlayAspect:      int(s.overlayAspect),
		ChromaKey:          s.chromaKey,
	}
	for fav, ok := range s.favorites {
		if ok {
			cfg.Favorites = append(cfg.Favorites, fav)
		}
	}
	sort.Strings(cfg.Favorites)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(settingsPath()), 0755)
	_ = os.WriteFile(settingsPath(), data, 0644)
}

func settingsPath() string {
	base := os.Getenv("APPDATA")
	if base == "" {
		base = "."
	}
	return filepath.Join(base, "MusicVisualizerPro", "settings.json")
}

type palette struct {
	background uintptr
	panel      uintptr
	text       uintptr
	dim        uintptr
	accent     uintptr
	accent2    uintptr
}

func currentPalette() palette {
	return paletteFor(app.theme)
}

func paletteFor(theme theme) palette {
	if theme >= themeCustomBase {
		idx := int(theme - themeCustomBase)
		if idx >= 0 && idx < len(app.customThemes) {
			return paletteFromCustom(app.customThemes[idx])
		}
	}
	switch theme {
	case themeLava:
		return palette{background: rgb(18, 8, 6), panel: rgb(58, 23, 18), text: rgb(255, 238, 220), dim: rgb(220, 145, 110), accent: rgb(255, 92, 40), accent2: rgb(255, 190, 40)}
	case themeCyberpunk:
		return palette{background: rgb(11, 8, 25), panel: rgb(36, 20, 65), text: rgb(245, 236, 255), dim: rgb(192, 135, 255), accent: rgb(255, 56, 188), accent2: rgb(55, 235, 255)}
	case themeOcean:
		return palette{background: rgb(4, 18, 28), panel: rgb(10, 48, 68), text: rgb(224, 248, 255), dim: rgb(112, 190, 215), accent: rgb(54, 210, 235), accent2: rgb(88, 245, 178)}
	default:
		return palette{background: rgb(10, 12, 22), panel: rgb(26, 32, 51), text: rgb(245, 247, 255), dim: rgb(170, 180, 205), accent: rgb(80, 210, 255), accent2: rgb(185, 90, 255)}
	}
}

func playLabel() string {
	if app.playing {
		return "Pause"
	}
	return "Play"
}

func inputLabel() string {
	if app.desktopMode {
		return "Input On"
	}
	return "Input"
}

func themeName(theme theme) string {
	if theme >= themeCustomBase {
		idx := int(theme - themeCustomBase)
		if idx >= 0 && idx < len(app.customThemes) {
			return app.customThemes[idx].Name
		}
		return "Custom"
	}
	switch theme {
	case themeLava:
		return "Lava"
	case themeCyberpunk:
		return "Cyberpunk"
	case themeOcean:
		return "Ocean"
	default:
		return "Neon"
	}
}

func repeatName(mode repeatMode) string {
	switch mode {
	case repeatOne:
		return "One"
	case repeatAll:
		return "All"
	default:
		return "Off"
	}
}

func sleepText(minutes int) string {
	if minutes == 0 {
		return "Off"
	}
	return fmt.Sprintf("%dm", minutes)
}

func pointInRect(x, y int32, r rect) bool {
	return x >= r.left && x <= r.right && y >= r.top && y <= r.bottom
}

func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

func hashPath(path string) uint32 {
	var hash uint32 = 2166136261
	for _, b := range []byte(strings.ToLower(path)) {
		hash ^= uint32(b)
		hash *= 16777619
	}
	return hash
}

func writeSnapshot(path string, bars []float64, mode visualMode, theme theme) error {
	const width = 900
	const height = 500
	palette := paletteFor(theme)
	bgR, bgG, bgB := colorParts(palette.background)
	accentR, accentG, accentB := colorParts(palette.accent)
	var out strings.Builder
	out.WriteString(fmt.Sprintf("P3\n%d %d\n255\n", width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b := bgR, bgG, bgB
			index := x * len(bars) / width
			barHeight := int(clamp(bars[index], 0, 1) * float64(height-80))
			switch mode {
			case modeHalo, modeTunnel, modeParticles:
				dx := float64(x - width/2)
				dy := float64(y - height/2)
				dist := math.Sqrt(dx*dx + dy*dy)
				angle := math.Atan2(dy, dx) + math.Pi
				idx := int(angle / (math.Pi * 2) * float64(len(bars)))
				limit := 80 + bars[idx%len(bars)]*180
				if dist > 90 && dist < limit {
					r, g, b = accentR, accentG, accentB
				}
			default:
				if y > height-40-barHeight && y < height-40 {
					r, g, b = accentR, accentG, accentB
				}
			}
			out.WriteString(fmt.Sprintf("%d %d %d ", r, g, b))
		}
		out.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(out.String()), 0644)
}

func colorParts(color uintptr) (int, int, int) {
	return int(color & 0xff), int((color >> 8) & 0xff), int((color >> 16) & 0xff)
}

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func (s *appState) seekFromPoint(x, y int32) bool {
	if s.filePath == "" || s.duration <= 0 {
		return false
	}
	var bounds rect
	procGetClientRect.Call(s.hwnd, uintptr(unsafe.Pointer(&bounds)))
	bar := progressRect(bounds.right-bounds.left, bounds.bottom-bounds.top)
	if y < bar.top-12 || y > bar.bottom+24 {
		return false
	}
	progress := clamp(float64(x-bar.left)/float64(bar.right-bar.left), 0, 1)
	pos := time.Duration(progress * float64(s.duration))
	s.dragPosition = pos
	s.seekTo(pos, s.playing)
	return true
}

func volumeSliderRect(width, height int32) rect {
	return volumeSliderRectFor(layoutFor(app.uiStyle, width, height), width)
}

func progressRect(width, height int32) rect {
	return layoutFor(app.uiStyle, width, height).progressBar(width)
}

func mousePoint(lParam uintptr) (int32, int32) {
	x := int16(lParam & 0xffff)
	y := int16((lParam >> 16) & 0xffff)
	return int32(x), int32(y)
}

func line(hdc uintptr, x1, y1, x2, y2 int32, color uintptr, width int32) {
	pen, _, _ := procCreatePen.Call(0, uintptr(width), color)
	if pen == 0 {
		return
	}
	old, _, _ := procSelectObject.Call(hdc, pen)
	procMoveToEx.Call(hdc, uintptr(x1), uintptr(y1), 0)
	procLineTo.Call(hdc, uintptr(x2), uintptr(y2))
	procSelectObject.Call(hdc, old)
	procDeleteObject.Call(pen)
}

func ellipse(hdc uintptr, left, top, right, bottom int32, color uintptr) {
	brush, _, _ := procCreateSolidBrush.Call(color)
	if brush == 0 {
		return
	}
	old, _, _ := procSelectObject.Call(hdc, brush)
	procEllipse.Call(hdc, uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	procSelectObject.Call(hdc, old)
	procDeleteObject.Call(brush)
}

func gradientColor(index, total int, alpha byte) uintptr {
	if total <= 1 {
		return rgb(80, 210, alpha)
	}
	t := float64(index) / float64(total-1)
	r := byte(60 + 170*math.Pow(t, 1.2))
	g := byte(210 - 90*t)
	b := byte(255 - 115*t)
	return rgb(r, g, b)
}

func bandGradientColor(index, total int, bars []float64) uintptr {
	bass, mid, treble := visual.SplitBands(bars)
	t := float64(index) / float64(max(1, total-1))
	r := byte(40 + bass*200 + t*30)
	g := byte(40 + mid*200 + t*20)
	b := byte(40 + treble*200 + (1-t)*40)
	return rgb(r, g, b)
}

func dimColor(color uintptr, factor float64) uintptr {
	r := byte(float64(color&0xff) * factor)
	g := byte(float64((color>>8)&0xff) * factor)
	b := byte(float64((color>>16)&0xff) * factor)
	return rgb(r, g, b)
}

func modeName(mode visualMode) string {
	switch mode {
	case modeMirror:
		return "Mirror"
	case modeBlocks:
		return "Blocks"
	case modeWave:
		return "Wave"
	case modeHalo:
		return "Halo"
	case modeSpectrum:
		return "Spectrum"
	case modeFire:
		return "Fire"
	case modeParticles:
		return "Particles"
	case modeTunnel:
		return "Tunnel"
	case modePlasma:
		return "Plasma"
	case modeAurora:
		return "Aurora"
	case modeMandala:
		return "Mandala"
	case modeStarfield:
		return "Starfield"
	case modeKaleidoscope:
		return "Kaleidoscope"
	case modeNeonCity:
		return "Neon City"
	case modeSupernova:
		return "Supernova"
	case modeLiquid:
		return "Liquid"
	case modeOrbit:
		return "Orbit"
	case modeWaveform3D:
		return "Waveform 3D"
	case modeFluid:
		return "Fluid"
	case modeGalaxy:
		return "Galaxy"
	case modeChromatic:
		return "Chromatic"
	case modeSpectrogram:
		return "Spectrogram"
	case modeOscilloscope:
		return "Oscilloscope"
	case modeLissajous:
		return "Lissajous"
	case modeVortex:
		return "Vortex"
	case modeComets:
		return "Comets"
	case modeBassDrop:
		return "Bass Drop"
	case modeRain:
		return "Rain"
	case modePrism:
		return "Prism"
	default:
		return "Classic"
	}
}

func openTrack(path string) error {
	closeTrack()
	if err := mciSend(fmt.Sprintf(`open "%s" alias musicvisualizer_track`, escapeMCI(path))); err != nil {
		return err
	}
	if err := mciSend("set musicvisualizer_track time format milliseconds"); err != nil {
		closeTrack()
		return err
	}
	return nil
}

func playTrackFrom(pos time.Duration) error {
	return mciSendNotify(fmt.Sprintf("play musicvisualizer_track from %d notify", durationMS(pos)), app.hwnd)
}

func pauseTrack() error {
	return mciSend("pause musicvisualizer_track")
}

func stopTrack() error {
	return mciSend("stop musicvisualizer_track")
}

func closeTrack() {
	_ = mciSend("stop musicvisualizer_track")
	_ = mciSend("close musicvisualizer_track")
}

func mciSend(command string) error {
	commandPtr, err := syscall.UTF16PtrFromString(command)
	if err != nil {
		return err
	}
	ret, _, _ := procMCISendStringW.Call(uintptr(unsafe.Pointer(commandPtr)), 0, 0, 0)
	if ret == 0 {
		return nil
	}
	return fmt.Errorf("%s", mciError(ret))
}

func mciSendNotify(command string, hwnd uintptr) error {
	commandPtr, err := syscall.UTF16PtrFromString(command)
	if err != nil {
		return err
	}
	ret, _, _ := procMCISendStringW.Call(uintptr(unsafe.Pointer(commandPtr)), 0, 0, hwnd)
	if ret == 0 {
		return nil
	}
	return fmt.Errorf("%s", mciError(ret))
}

func trackDuration() (time.Duration, error) {
	var buffer [64]uint16
	command, _ := syscall.UTF16PtrFromString("status musicvisualizer_track length")
	ret, _, _ := procMCISendStringW.Call(uintptr(unsafe.Pointer(command)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0)
	if ret != 0 {
		return 0, fmt.Errorf("%s", mciError(ret))
	}
	var ms int64
	_, err := fmt.Sscanf(syscall.UTF16ToString(buffer[:]), "%d", &ms)
	if err != nil {
		return 0, err
	}
	return time.Duration(ms) * time.Millisecond, nil
}

func mciError(code uintptr) string {
	var buffer [256]uint16
	ret, _, _ := procMCIGetErrorStringW.Call(code, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if ret == 0 {
		return fmt.Sprintf("MCI error %d", code)
	}
	return syscall.UTF16ToString(buffer[:])
}

func escapeMCI(path string) string {
	return strings.ReplaceAll(path, `"`, `""`)
}

func invalidate() {
	if app.hwnd != 0 {
		procInvalidateRect.Call(app.hwnd, 0, 0)
	}
}

func idleBars() []float64 {
	bars := make([]float64, barCount)
	now := float64(time.Now().UnixMilli()) / 280
	for i := range bars {
		wave := math.Sin(now+float64(i)*0.45)*0.5 + 0.5
		bars[i] = 0.1 + wave*0.18
	}
	return bars
}

func fill(hdc uintptr, r rect, color uintptr) {
	brush, _, _ := procCreateSolidBrush.Call(color)
	if brush == 0 {
		return
	}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
	procDeleteObject.Call(brush)
}

func textOut(hdc uintptr, x, y int32, text string) {
	chars, err := syscall.UTF16FromString(text)
	if err != nil || len(chars) == 0 {
		return
	}
	procTextOutW.Call(hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&chars[0])), uintptr(len(chars)-1))
}

func showMessage(hwnd uintptr, title, body string, flags uintptr) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	bodyPtr, _ := syscall.UTF16PtrFromString(body)
	procMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(bodyPtr)), uintptr(unsafe.Pointer(titlePtr)), flags)
}

func durationText(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	totalSeconds := int(duration.Seconds() + 0.5)
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func durationMS(duration time.Duration) int64 {
	if duration < 0 {
		return 0
	}
	return int64(duration / time.Millisecond)
}

func utf16WithNULs(s string) []uint16 {
	if !strings.HasSuffix(s, "\x00") {
		s += "\x00"
	}
	return utf16.Encode([]rune(s))
}

func rgb(r, g, b byte) uintptr {
	return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func clampDuration(value, low, high time.Duration) time.Duration {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
