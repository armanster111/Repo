//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/armanster111/music-visualizer/internal/visual"
)

const (
	coinitMultithreaded  = 0x0
	rpcEChangedMode      = 0x80010106
	clsctxAll            = 0x17
	eRender              = 0
	eConsole             = 0
	audioShareModeShared = 0
	audioStreamLoopback  = 0x00020000
	audioBufferSilent    = 0x00000002
	waveFormatPCM        = 0x0001
	waveFormatIEEEFloat  = 0x0003
	waveFormatExtensible = 0xfffe
	loopbackBuffer100NS  = 10000000
)

var (
	ole32 = syscall.NewLazyDLL("ole32.dll")

	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree    = ole32.NewProc("CoTaskMemFree")

	clsidMMDeviceEnumerator = guid{0xbcde0395, 0xe52f, 0x467c, [8]byte{0x8e, 0x3d, 0xc4, 0x57, 0x92, 0x91, 0x69, 0x2e}}
	iidIMMDeviceEnumerator  = guid{0xa95664d2, 0x9614, 0x4f35, [8]byte{0xa7, 0x46, 0xde, 0x8d, 0xb6, 0x36, 0x17, 0xe6}}
	iidIAudioClient         = guid{0x1cb9ad4c, 0xdbfa, 0x4c32, [8]byte{0xb1, 0x78, 0xc2, 0xf5, 0x68, 0xa7, 0x03, 0xb2}}
	iidIAudioCaptureClient  = guid{0xc8adbd64, 0xe71e, 0x48a0, [8]byte{0xa4, 0xde, 0x18, 0x5c, 0x39, 0x5c, 0xd3, 0x17}}
	subtypePCM              = guid{0x00000001, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}}
	subtypeIEEEFloat        = guid{0x00000003, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}}
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type waveFormat struct {
	formatTag      uint16
	channels       uint16
	samplesPerSec  uint32
	avgBytesPerSec uint32
	blockAlign     uint16
	bitsPerSample  uint16
	cbSize         uint16
	subFormat      guid
}

type desktopInput struct {
	mu          sync.RWMutex
	bars        []float64
	ring        []float64
	sampleRate  int
	active      bool
	lastErr     error
	stop        chan struct{}
	sensitivity float64
	deviceIndex int
}

func newDesktopInput() *desktopInput {
	return &desktopInput{bars: make([]float64, barCount), sensitivity: 1.0}
}

func (d *desktopInput) setSensitivity(v float64) {
	d.mu.Lock()
	d.sensitivity = v
	d.mu.Unlock()
}

func (d *desktopInput) setDeviceIndex(idx int) error {
	d.mu.Lock()
	d.deviceIndex = idx
	d.mu.Unlock()
	return nil
}

func (d *desktopInput) start() error {
	d.mu.Lock()
	if d.active {
		d.mu.Unlock()
		return nil
	}
	d.stop = make(chan struct{})
	d.active = true
	d.lastErr = nil
	d.bars = make([]float64, barCount)
	stop := d.stop
	d.mu.Unlock()

	go d.captureLoop(stop)
	return nil
}

func (d *desktopInput) stopCapture() {
	d.mu.Lock()
	if d.active && d.stop != nil {
		close(d.stop)
	}
	d.active = false
	d.stop = nil
	d.mu.Unlock()
}

func (d *desktopInput) isActive() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.active
}

func (d *desktopInput) snapshot() []float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if !d.active || len(d.bars) == 0 {
		return nil
	}
	out := make([]float64, len(d.bars))
	copy(out, d.bars)
	return out
}

func (d *desktopInput) errText() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.lastErr == nil {
		return ""
	}
	return d.lastErr.Error()
}

func (d *desktopInput) captureLoop(stop <-chan struct{}) {
	if err := d.runLoopback(stop); err != nil {
		d.mu.Lock()
		d.lastErr = err
		d.active = false
		d.mu.Unlock()
	}
}

func (d *desktopInput) runLoopback(stop <-chan struct{}) error {
	hr, _, _ := procCoInitializeEx.Call(0, coinitMultithreaded)
	coInitialized := true
	if failedHRESULT(hr) {
		if uint32(hr) != rpcEChangedMode {
			return hresultError("CoInitializeEx", hr)
		}
		coInitialized = false
	}
	if coInitialized {
		defer procCoUninitialize.Call()
	}

	var enumerator uintptr
	hr, _, _ = procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidMMDeviceEnumerator)),
		0,
		clsctxAll,
		uintptr(unsafe.Pointer(&iidIMMDeviceEnumerator)),
		uintptr(unsafe.Pointer(&enumerator)),
	)
	if failedHRESULT(hr) {
		return hresultError("CoCreateInstance IMMDeviceEnumerator", hr)
	}
	defer comRelease(enumerator)

	var device uintptr
	hr, _, _ = comCall(enumerator, 4, eRender, eConsole, uintptr(unsafe.Pointer(&device)))
	if failedHRESULT(hr) {
		return hresultError("GetDefaultAudioEndpoint", hr)
	}
	defer comRelease(device)

	var audioClient uintptr
	hr, _, _ = comCall(device, 3, uintptr(unsafe.Pointer(&iidIAudioClient)), clsctxAll, 0, uintptr(unsafe.Pointer(&audioClient)))
	if failedHRESULT(hr) {
		return hresultError("Activate IAudioClient", hr)
	}
	defer comRelease(audioClient)

	var formatPtr uintptr
	hr, _, _ = comCall(audioClient, 8, uintptr(unsafe.Pointer(&formatPtr)))
	if failedHRESULT(hr) {
		return hresultError("GetMixFormat", hr)
	}
	defer procCoTaskMemFree.Call(formatPtr)
	format := parseWaveFormat(formatPtr)
	d.mu.Lock()
	d.sampleRate = int(format.samplesPerSec)
	d.mu.Unlock()

	hr, _, _ = comCall(audioClient, 3, audioShareModeShared, audioStreamLoopback, loopbackBuffer100NS, 0, formatPtr, 0)
	if failedHRESULT(hr) {
		return hresultError("IAudioClient Initialize loopback", hr)
	}

	var captureClient uintptr
	hr, _, _ = comCall(audioClient, 14, uintptr(unsafe.Pointer(&iidIAudioCaptureClient)), uintptr(unsafe.Pointer(&captureClient)))
	if failedHRESULT(hr) {
		return hresultError("GetService IAudioCaptureClient", hr)
	}
	defer comRelease(captureClient)

	hr, _, _ = comCall(audioClient, 10)
	if failedHRESULT(hr) {
		return hresultError("IAudioClient Start", hr)
	}
	defer comCall(audioClient, 11)

	for {
		select {
		case <-stop:
			return nil
		default:
		}

		var packetFrames uint32
		hr, _, _ = comCall(captureClient, 5, uintptr(unsafe.Pointer(&packetFrames)))
		if failedHRESULT(hr) {
			return hresultError("GetNextPacketSize", hr)
		}

		for packetFrames > 0 {
			var data uintptr
			var frames uint32
			var flags uint32
			hr, _, _ = comCall(
				captureClient,
				3,
				uintptr(unsafe.Pointer(&data)),
				uintptr(unsafe.Pointer(&frames)),
				uintptr(unsafe.Pointer(&flags)),
				0,
				0,
			)
			if failedHRESULT(hr) {
				return hresultError("GetBuffer", hr)
			}

			if flags&audioBufferSilent != 0 {
				d.updateBars(nil)
			} else {
				d.updateBars(samplesFromBuffer(data, int(frames), format))
			}

			hr, _, _ = comCall(captureClient, 4, uintptr(frames))
			if failedHRESULT(hr) {
				return hresultError("ReleaseBuffer", hr)
			}

			hr, _, _ = comCall(captureClient, 5, uintptr(unsafe.Pointer(&packetFrames)))
			if failedHRESULT(hr) {
				return hresultError("GetNextPacketSize", hr)
			}
		}

		time.Sleep(15 * time.Millisecond)
	}
}

func (d *desktopInput) updateBars(samples []float64) {
	target := make([]float64, barCount)
	if len(samples) > 0 {
		d.mu.Lock()
		d.ring = append(d.ring, samples...)
		if len(d.ring) > 8192 {
			d.ring = append([]float64(nil), d.ring[len(d.ring)-8192:]...)
		}
		sr := d.sampleRate
		ring := append([]float64(nil), d.ring...)
		d.mu.Unlock()
		if sr <= 0 {
			sr = 44100
		}
		windowSize := 2048
		if len(ring) >= 512 {
			window := ring
			if len(window) > windowSize {
				window = window[len(window)-windowSize:]
			}
			padded := make([]float64, windowSize)
			copy(padded[windowSize-len(window):], window)
			if fftBars := visual.AnalyzeWindow(padded, sr, barCount); fftBars != nil {
				for i, v := range fftBars {
					target[i] = math.Min(1, v*3.0*d.sensitivity)
				}
			}
		}
		normalizeLiveBars(target)
	}

	d.mu.Lock()
	for i := range d.bars {
		d.bars[i] += (target[i] - d.bars[i]) * 0.45
		if d.bars[i] < 0.015 {
			d.bars[i] = 0
		}
	}
	d.mu.Unlock()
}

func samplesFromBuffer(data uintptr, frames int, format waveFormat) []float64 {
	if data == 0 || frames <= 0 || format.channels == 0 || format.blockAlign == 0 {
		return nil
	}
	bytesPerSample := int(format.bitsPerSample / 8)
	if bytesPerSample <= 0 {
		return nil
	}
	out := make([]float64, 0, frames)
	blockAlign := int(format.blockAlign)
	channels := int(format.channels)
	totalBytes := frames * blockAlign
	raw := unsafe.Slice((*byte)(unsafe.Pointer(data)), totalBytes)

	for frame := 0; frame < frames; frame++ {
		var sum float64
		for ch := 0; ch < channels; ch++ {
			offset := frame*blockAlign + ch*bytesPerSample
			sum += sampleValue(raw[offset:offset+bytesPerSample], format)
		}
		out = append(out, sum/float64(channels))
	}
	return out
}

func sampleValue(data []byte, format waveFormat) float64 {
	if format.isFloat() {
		if len(data) >= 4 {
			return float64(math.Float32frombits(binary.LittleEndian.Uint32(data)))
		}
		return 0
	}

	switch format.bitsPerSample {
	case 8:
		return (float64(data[0]) - 128) / 128
	case 16:
		return float64(int16(binary.LittleEndian.Uint16(data))) / float64(math.MaxInt16)
	case 24:
		v := int32(data[0]) | int32(data[1])<<8 | int32(data[2])<<16
		if v&0x800000 != 0 {
			v |= ^0xffffff
		}
		return float64(v) / 8388607
	case 32:
		return float64(int32(binary.LittleEndian.Uint32(data))) / float64(math.MaxInt32)
	default:
		return 0
	}
}

func (f waveFormat) isFloat() bool {
	return f.formatTag == waveFormatIEEEFloat || (f.formatTag == waveFormatExtensible && f.subFormat == subtypeIEEEFloat)
}

func parseWaveFormat(ptr uintptr) waveFormat {
	format := waveFormat{
		formatTag:      *(*uint16)(unsafe.Pointer(ptr)),
		channels:       *(*uint16)(unsafe.Pointer(ptr + 2)),
		samplesPerSec:  *(*uint32)(unsafe.Pointer(ptr + 4)),
		avgBytesPerSec: *(*uint32)(unsafe.Pointer(ptr + 8)),
		blockAlign:     *(*uint16)(unsafe.Pointer(ptr + 12)),
		bitsPerSample:  *(*uint16)(unsafe.Pointer(ptr + 14)),
		cbSize:         *(*uint16)(unsafe.Pointer(ptr + 16)),
	}
	if format.formatTag == waveFormatExtensible && format.cbSize >= 22 {
		format.subFormat = *(*guid)(unsafe.Pointer(ptr + 24))
	} else if format.formatTag == waveFormatIEEEFloat {
		format.subFormat = subtypeIEEEFloat
	} else if format.formatTag == waveFormatPCM {
		format.subFormat = subtypePCM
	}
	return format
}

func normalizeLiveBars(bars []float64) {
	var peak float64
	for _, value := range bars {
		peak = math.Max(peak, value)
	}
	if peak < 0.04 {
		return
	}
	for i, value := range bars {
		bars[i] = math.Min(1, value/peak)
	}
}

func comCall(obj uintptr, method int, args ...uintptr) (uintptr, uintptr, syscall.Errno) {
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	fn := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(method)*unsafe.Sizeof(uintptr(0))))
	callArgs := make([]uintptr, 0, len(args)+1)
	callArgs = append(callArgs, obj)
	callArgs = append(callArgs, args...)
	r1, r2, err := syscall.SyscallN(fn, callArgs...)
	return r1, r2, err
}

func comRelease(obj uintptr) {
	if obj != 0 {
		comCall(obj, 2)
	}
}

func failedHRESULT(hr uintptr) bool {
	return uint32(hr)&0x80000000 != 0
}

func hresultError(label string, hr uintptr) error {
	return fmt.Errorf("%s failed with HRESULT 0x%08x", label, uint32(hr))
}
