package lyrics

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var timeTag = regexp.MustCompile(`^\[(\d+):(\d+(?:\.\d+)?)\]`)

// Line is one timed lyric line.
type Line struct {
	At   time.Duration
	Text string
}

// Load reads a sidecar .lrc file if present.
func Load(audioPath string) []Line {
	lrcPath := strings.TrimSuffix(audioPath, filepath.Ext(audioPath)) + ".lrc"
	f, err := os.Open(lrcPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []Line
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		matches := timeTag.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		minutes, _ := strconv.Atoi(matches[1])
		seconds, _ := strconv.ParseFloat(matches[2], 64)
		text := strings.TrimSpace(timeTag.ReplaceAllString(line, ""))
		if text == "" {
			continue
		}
		lines = append(lines, Line{
			At:   time.Duration(minutes)*time.Minute + time.Duration(seconds*float64(time.Second)),
			Text: text,
		})
	}
	return lines
}

// At returns the active lyric line for a playback position.
func At(lines []Line, pos time.Duration) string {
	if len(lines) == 0 {
		return ""
	}
	active := ""
	for _, line := range lines {
		if line.At <= pos {
			active = line.Text
			continue
		}
		break
	}
	return active
}
