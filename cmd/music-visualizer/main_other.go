//go:build !windows

package main

import "fmt"

func main() {
	fmt.Println("Music Visualizer is a Windows desktop app.")
	fmt.Println("Build it with: GOOS=windows GOARCH=amd64 go build -ldflags='-H windowsgui' -o dist/MusicVisualizerPro.exe ./cmd/music-visualizer")
}
