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

	wmDestroy = 0x0002
	wmPaint   = 0x000f
	wmTimer   = 0x0113
	wmKeyDown = 0x0100
	wmCreate  = 0x0001

	vkEscape = 0x1b
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
	procSetTimer         = user32.NewProc("SetTimer")
	procShowWindow       = user32.NewProc("ShowWindow")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procUpdateWindow     = user32.NewProc("UpdateWindow")

	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procTextOutW         = gdi32.NewProc("TextOutW")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	procPlaySoundW       = winmm.NewProc("PlaySoundW")
)

type appState struct {
	hwnd        uintptr
	filePath    string
	status      string
	frames      []visual.Frame
	currentBars []float64
	duration    time.Duration
	startedAt   time.Time
	playing     bool
}

var app = &appState{
	status: "Press O to open a WAV file. Space restarts playback. Esc exits.",
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
		case vkSpace:
			app.restart()
		case uintptr('O'):
			if path, ok := openWAVDialog(hwnd); ok {
				app.load(path)
			}
		}
		return 0
	case wmPaint:
		draw(hwnd)
		return 0
	case wmDestroy:
		stopSound()
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
		stopSound()
		invalidate()
		return
	}

	frames := visual.BuildFrames(audio.Mono, audio.SampleRate, fps, barCount)
	if len(frames) == 0 {
		s.status = "Loaded WAV file has no audio samples."
		s.frames = nil
		s.playing = false
		stopSound()
		invalidate()
		return
	}

	s.filePath = path
	s.frames = frames
	s.duration = time.Duration(audio.DurationSeconds() * float64(time.Second))
	s.status = fmt.Sprintf("Playing %s (%s, %d Hz, %d-bit)", filepath.Base(path), durationText(s.duration), audio.SampleRate, audio.BitsPerSample)
	s.restart()
}

func (s *appState) restart() {
	if s.filePath == "" {
		return
	}
	stopSound()
	if !playSound(s.filePath) {
		s.status = "Windows could not play this WAV file."
		s.playing = false
		invalidate()
		return
	}
	s.startedAt = time.Now()
	s.playing = true
	invalidate()
}

func (s *appState) targetBars() []float64 {
	if len(s.frames) == 0 {
		return idleBars()
	}
	if !s.playing {
		return s.frames[0].Bars
	}

	elapsed := time.Since(s.startedAt)
	if elapsed >= s.duration {
		s.playing = false
		stopSound()
		return s.frames[len(s.frames)-1].Bars
	}

	index := int(elapsed.Seconds() * fps)
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
	textOut(hdc, 28, 24, appTitle)
	procSetTextColor.Call(hdc, rgb(170, 180, 205))
	textOut(hdc, 28, 52, app.status)
	textOut(hdc, 28, height-30, "O: open WAV    Space: restart playback    Esc: quit")

	bars := app.targetBars()
	if len(app.currentBars) != len(bars) {
		app.currentBars = make([]float64, len(bars))
	}

	top := int32(112)
	bottom := height - 58
	if bottom <= top {
		return
	}
	usableHeight := bottom - top
	margin := int32(28)
	gap := int32(5)
	barWidth := (width - margin*2 - gap*int32(len(bars)-1)) / int32(len(bars))
	if barWidth < 2 {
		barWidth = 2
		gap = 2
	}

	for i, target := range bars {
		app.currentBars[i] += (target - app.currentBars[i]) * 0.35
		value := math.Pow(clamp(app.currentBars[i], 0, 1), 0.75)
		barHeight := int32(value * float64(usableHeight))
		left := margin + int32(i)*(barWidth+gap)
		right := left + barWidth
		color := rgb(byte(40+i*4), byte(160+i*2), byte(255-i*3))
		fill(hdc, rect{left: left, top: bottom - barHeight, right: right, bottom: bottom}, color)
	}
}

func playSound(path string) bool {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	ret, _, _ := procPlaySoundW.Call(uintptr(unsafe.Pointer(pathPtr)), 0, sndFilename|sndAsync|sndNoDefault)
	return ret != 0
}

func stopSound() {
	procPlaySoundW.Call(0, 0, 0)
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
	totalSeconds := int(duration.Seconds() + 0.5)
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
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
