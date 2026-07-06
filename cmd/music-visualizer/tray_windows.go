//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	nimAdd    = 0x00000000
	nimDelete = 0x00000002
	nimModify = 0x00000001

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	wmTrayIcon = wmApp + 2

	idiApplication = 32512

	wmRButtonUp = 0x0205
)

var (
	shell32Tray          = syscall.NewLazyDLL("shell32.dll")
	procShellNotifyIconW = shell32Tray.NewProc("Shell_NotifyIconW")
)

type notifyIconData struct {
	size            uint32
	hwnd            uintptr
	id              uint32
	flags           uint32
	callbackMessage uint32
	hIcon           uintptr
	tip             [128]uint16
	state           uint32
	stateMask       uint32
	info            [256]uint16
	version         uint32
	infoFlags       uint32
	guidItem        [16]byte
	hBalloonIcon    uintptr
}

func (s *appState) initTray() {
	if s.hwnd == 0 {
		return
	}
	icon, _, _ := procLoadIconW.Call(0, idiApplication)
	tip, _ := syscall.UTF16FromString(appTitle + " Pro")
	data := notifyIconData{
		size:            uint32(unsafe.Sizeof(notifyIconData{})),
		hwnd:            s.hwnd,
		id:              1,
		flags:           nifMessage | nifIcon | nifTip,
		callbackMessage: wmTrayIcon,
		hIcon:           icon,
	}
	copy(data.tip[:], tip)
	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&data)))
}

func (s *appState) removeTray() {
	if s.hwnd == 0 {
		return
	}
	data := notifyIconData{
		size: uint32(unsafe.Sizeof(notifyIconData{})),
		hwnd: s.hwnd,
		id:   1,
	}
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&data)))
}

func (s *appState) handleTrayMessage(lParam uintptr) {
	switch lParam {
	case wmLButtonUp:
		procShowWindow.Call(s.hwnd, swShowNormal)
		procSetForegroundWindow.Call(s.hwnd)
	case wmRButtonUp:
		s.togglePlay()
	}
}
