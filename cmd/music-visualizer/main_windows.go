//go:build windows

package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/armanster111/music-visualizer/internal/visual"
	"github.com/armanster111/music-visualizer/internal/wav"
)

const (
	appTitle = "Music Visualizer"
	fps      = 30
	barCount = 48

	csHRedraw = 0x0002
	csVRedraw = 0x0001

	cwUseDefault       = 0x80000000
	wsOverlappedWindow = 0x00cf0000

	swShowDefault = 10

	wmDestroy     = 0x0002
	wmPaint       = 0x000f
	wmTimer       = 0x0113
	wmKeyDown     = 0x0100
	wmCreate      = 0x0001
	wmMouseMove   = 0x0200
	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202

	mkLButton = 0x0001

	vkEscape = 0x1b
	vkHome   = 0x24
	vkLeft   = 0x25
	vkRight  = 0x27
	vkSpace  = 0x20

	idcArrow = 32512

	timerID = 1

	transparent = 1

	ofnPathMustExist = 0x00000800
	ofnFileMustExist = 0x00001000
	ofnExplorer      = 0x00080000

	sndAsync     = 0x0001
	sndNoDefault = 0x0002
	sndFilename  = 0x00020000

	mbIconError = 0x00000010
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	winmm    = syscall.NewLazyDLL("winmm.dll")

	procBeginPaint       = user32.NewProc("BeginPaint")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procEndPaint         = user32.NewProc("EndPaint")
	procFillRect         = user32.NewProc("FillRect")
	procGetClientRect    = user32.NewProc("GetClientRect")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procInvalidateRect   = user32.NewProc("InvalidateRect")
	procLoadCursorW      = user32.NewProc("LoadCursorW")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procReleaseCapture   = user32.NewProc("ReleaseCapture")
	procSetCapture       = user32.NewProc("SetCapture")
	procSetTimer         = user32.NewProc("SetTimer")
	procShowWindow       = user32.NewProc("ShowWindow")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procUpdateWindow     = user32.NewProc("UpdateWindow")

	procCreatePen        = gdi32.NewProc("CreatePen")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
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
	procMCIGetErrorStringW = winmm.NewProc("mciGetErrorStringW")
	procMCISendStringW     = winmm.NewProc("mciSendStringW")
	procPlaySoundW         = winmm.NewProc("PlaySoundW")
)

type visualMode int

const (
	modeClassic visualMode = iota
	modeMirror
	modeBlocks
	modeWave
	modeHalo
	modeCount
)

type appState struct {
	hwnd           uintptr
	filePath       string
	status         string
	frames         []visual.Frame
	currentBars    []float64
	duration       time.Duration
	startedAt      time.Time
	playbackOffset time.Duration
	dragPosition   time.Duration
	mode           visualMode
	seekLarge      bool
	playing        bool
	paused         bool
	dragging       bool
}

var app = &appState{
	status: "Press O to open a WAV file. V changes visualizers. Click the track bar to seek.",
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

	instance, _, _ := procGetModuleHandleW.Call(0)
	className, _ := syscall.UTF16PtrFromString("MusicVisualizerWindow")
	title, _ := syscall.UTF16PtrFromString(appTitle)
	cursor, _, _ := procLoadCursorW.Call(0, idcArrow)

	wc := wndClassEx{
		size:      uint32(unsafe.Sizeof(wndClassEx{})),
		style:     csHRedraw | csVRedraw,
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

	procShowWindow.Call(hwnd, swShowDefault)
	procUpdateWindow.Call(hwnd)

	if len(os.Args) > 1 {
		app.load(os.Args[1])
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
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0
	case wmKeyDown:
		switch wParam {
		case vkEscape:
			procDestroyWindow.Call(hwnd)
		case vkHome:
			app.seekTo(0, app.playing)
		case vkLeft:
			app.seekBy(-app.seekStep())
		case vkRight:
			app.seekBy(app.seekStep())
		case vkSpace:
			app.togglePlay()
		case uintptr('O'):
			if path, ok := openWAVDialog(hwnd); ok {
				app.load(path)
			}
		case uintptr('R'):
			app.restart()
		case uintptr('T'):
			app.seekLarge = !app.seekLarge
			app.updateStatus()
			invalidate()
		case uintptr('V'):
			app.cycleMode()
		}
		return 0
	case wmLButtonDown:
		x, y := mousePoint(lParam)
		if app.seekFromPoint(x, y) {
			app.dragging = true
			procSetCapture.Call(hwnd)
		}
		return 0
	case wmMouseMove:
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
	case wmPaint:
		draw(hwnd)
		return 0
	case wmDestroy:
		closeTrack()
		procPostQuitMessage.Call(0)
		return 0
	default:
		ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
		return ret
	}
}

func (s *appState) load(path string) {
	audio, err := wav.DecodeFile(path)
	if err != nil {
		s.status = "Could not load WAV: " + err.Error()
		s.frames = nil
		s.playing = false
		closeTrack()
		invalidate()
		return
	}

	frames := visual.BuildFrames(audio.Mono, audio.SampleRate, fps, barCount)
	if len(frames) == 0 {
		s.status = "Loaded WAV file has no audio samples."
		s.frames = nil
		s.playing = false
		closeTrack()
		invalidate()
		return
	}

	if err := openTrack(path); err != nil {
		s.status = "Windows could not open this WAV file: " + err.Error()
		s.frames = nil
		s.playing = false
		closeTrack()
		invalidate()
		return
	}

	s.filePath = path
	s.frames = frames
	s.duration = time.Duration(audio.DurationSeconds() * float64(time.Second))
	s.playbackOffset = 0
	s.paused = false
	s.updateStatus()
	s.restart()
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
	} else if !s.paused {
		_ = stopTrack()
	}
	s.updateStatus()
}

func (s *appState) cycleMode() {
	s.mode = (s.mode + 1) % modeCount
	s.updateStatus()
	invalidate()
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
		s.status = "Press O to open a WAV file. V changes visualizers. Click the track bar to seek."
		return
	}
	state := "Playing"
	if s.paused {
		state = "Paused"
	} else if !s.playing {
		state = "Ready"
	}
	s.status = fmt.Sprintf("%s %s  %s/%s  Mode: %s  Seek: %s",
		state,
		filepath.Base(s.filePath),
		durationText(s.position()),
		durationText(s.duration),
		modeName(s.mode),
		durationText(s.seekStep()),
	)
}

func (s *appState) targetBars() []float64 {
	if len(s.frames) == 0 {
		return idleBars()
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

func openWAVDialog(hwnd uintptr) (string, bool) {
	var fileBuffer [4096]uint16
	filter := utf16WithNULs("WAV files (*.wav)\x00*.wav\x00All files (*.*)\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("Open a WAV file")

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
	fill(hdc, bounds, rgb(10, 12, 22))

	procSetBkMode.Call(hdc, transparent)
	procSetTextColor.Call(hdc, rgb(245, 247, 255))
	textOut(hdc, 28, 24, appTitle+" Pro")
	procSetTextColor.Call(hdc, rgb(170, 180, 205))
	textOut(hdc, 28, 52, app.status)
	textOut(hdc, 28, height-30, "O: open    Space: pause/play    V: visualizer    T: seek size    Left/Right: seek    R: restart    Esc: quit")
	drawProgress(hdc, width, height)

	bars := app.targetBars()
	if len(app.currentBars) != len(bars) {
		app.currentBars = make([]float64, len(bars))
	}

	top := int32(112)
	bottom := height - 104
	if bottom <= top {
		return
	}

	for i, target := range bars {
		app.currentBars[i] += (target - app.currentBars[i]) * 0.35
	}
	visualBounds := rect{left: 28, top: top, right: width - 28, bottom: bottom}
	drawVisualization(hdc, visualBounds, app.currentBars, app.mode)
}

func drawVisualization(hdc uintptr, bounds rect, bars []float64, mode visualMode) {
	switch mode {
	case modeMirror:
		drawMirrorBars(hdc, bounds, bars)
	case modeBlocks:
		drawBlockStack(hdc, bounds, bars)
	case modeWave:
		drawWaveLine(hdc, bounds, bars)
	case modeHalo:
		drawHalo(hdc, bounds, bars)
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
		color := gradientColor(i, len(bars), 255)
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

func drawProgress(hdc uintptr, width, height int32) {
	bar := progressRect(width, height)
	fill(hdc, bar, rgb(26, 32, 51))
	progress := app.progress()
	fill(hdc, rect{left: bar.left, top: bar.top, right: bar.left + int32(progress*float64(bar.right-bar.left)), bottom: bar.bottom}, rgb(80, 210, 255))
	knobX := bar.left + int32(progress*float64(bar.right-bar.left))
	fill(hdc, rect{left: knobX - 4, top: bar.top - 5, right: knobX + 4, bottom: bar.bottom + 5}, rgb(245, 247, 255))
	procSetTextColor.Call(hdc, rgb(170, 180, 205))
	textOut(hdc, bar.left, bar.bottom+14, "Click or drag this bar to shift through the track")
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

func progressRect(width, height int32) rect {
	return rect{
		left:   28,
		top:    height - 82,
		right:  width - 28,
		bottom: height - 68,
	}
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
	default:
		return "Classic"
	}
}

func openTrack(path string) error {
	closeTrack()
	if err := mciSend(fmt.Sprintf(`open "%s" type waveaudio alias musicvisualizer_track`, escapeMCI(path))); err != nil {
		return err
	}
	if err := mciSend("set musicvisualizer_track time format milliseconds"); err != nil {
		closeTrack()
		return err
	}
	return nil
}

func playTrackFrom(pos time.Duration) error {
	return mciSend(fmt.Sprintf("play musicvisualizer_track from %d", durationMS(pos)))
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
		procInvalidateRect.Call(app.hwnd, 0, 1)
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
