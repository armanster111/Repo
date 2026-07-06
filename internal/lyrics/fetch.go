package lyrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const lrclibSearchURL = "https://lrclib.net/api/search"

type lrclibResult struct {
	SyncedLyrics string `json:"syncedLyrics"`
	PlainLyrics  string `json:"plainLyrics"`
}

// FetchOnline searches LRCLIB for synced lyrics.
func FetchOnline(artist, title string) ([]Line, error) {
	artist = strings.TrimSpace(artist)
	title = strings.TrimSpace(title)
	if artist == "" && title == "" {
		return nil, fmt.Errorf("missing artist and title")
	}
	q := url.Values{}
	if title != "" {
		q.Set("track_name", title)
	}
	if artist != "" {
		q.Set("artist_name", artist)
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(lrclibSearchURL + "?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lyrics api status %d", resp.StatusCode)
	}
	var results []lrclibResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no lyrics found")
	}
	text := results[0].SyncedLyrics
	if strings.TrimSpace(text) == "" {
		text = results[0].PlainLyrics
	}
	return ParseLRC(text), nil
}

// ParseLRC parses synced or plain LRC text into timed lines.
func ParseLRC(text string) []Line {
	var lines []Line
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			matches := timeTag.FindStringSubmatch(line)
			if len(matches) == 3 {
				minutes, _ := strconv.Atoi(matches[1])
				seconds, _ := strconv.ParseFloat(matches[2], 64)
				body := strings.TrimSpace(timeTag.ReplaceAllString(line, ""))
				if body != "" {
					lines = append(lines, Line{
						At:   time.Duration(minutes)*time.Minute + time.Duration(seconds*float64(time.Second)),
						Text: body,
					})
				}
				continue
			}
		}
		lines = append(lines, Line{Text: line})
	}
	return lines
}
