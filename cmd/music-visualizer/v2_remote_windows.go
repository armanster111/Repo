//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type remoteState struct {
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Playing  bool    `json:"playing"`
	Position float64 `json:"position_seconds"`
	Duration float64 `json:"duration_seconds"`
	Mode     string  `json:"mode"`
	Mood     string  `json:"mood"`
	BPM      float64 `json:"bpm"`
}

var remoteMu sync.Mutex
var remoteServer *http.Server

func (s *appState) startRemoteControl() {
	if remoteServer != nil {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/nowplaying", func(w http.ResponseWriter, r *http.Request) {
		st := s.remoteSnapshot()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(st)
	})
	mux.HandleFunc("/api/play", func(w http.ResponseWriter, r *http.Request) {
		s.togglePlay()
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/pause", func(w http.ResponseWriter, r *http.Request) {
		if s.playing {
			s.togglePlay()
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/next", func(w http.ResponseWriter, r *http.Request) {
		s.nextTrack()
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/prev", func(w http.ResponseWriter, r *http.Request) {
		s.previousTrack()
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		st := s.remoteSnapshot()
		fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Music Visualizer Remote</title>
<style>body{font-family:Segoe UI,sans-serif;background:#120a22;color:#fff;padding:24px}
button{margin:6px;padding:10px 16px;font-size:16px}</style></head><body>
<h1>Music Visualizer Pro Remote</h1>
<p><b>%s</b> — %s</p>
<p>%s | Mood: %s | BPM: %.0f</p>
<button onclick="fetch('/api/prev',{method:'POST'})">Prev</button>
<button onclick="fetch('/api/play',{method:'POST'})">Play/Pause</button>
<button onclick="fetch('/api/next',{method:'POST'})">Next</button>
<script>setInterval(()=>location.reload(),3000)</script></body></html>`,
			st.Title, st.Artist, modeName(s.mode), st.Mood, st.BPM)
	})
	remoteServer = &http.Server{Addr: "127.0.0.1:8765", Handler: mux}
	s.remoteEnabled = true
	go func() {
		if err := remoteServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.remoteEnabled = false
		}
	}()
}

func (s *appState) stopRemoteControl() {
	if remoteServer != nil {
		_ = remoteServer.Close()
		remoteServer = nil
	}
	s.remoteEnabled = false
}

func (s *appState) remoteSnapshot() remoteState {
	return remoteState{
		Title:    s.meta.Title,
		Artist:   s.meta.Artist,
		Playing:  s.playing,
		Position: s.position().Seconds(),
		Duration: s.duration.Seconds(),
		Mode:     modeName(s.mode),
		Mood:     s.mood,
		BPM:      s.bpm,
	}
}

func (s *appState) exportSyncBundle() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Desktop", "music-visualizer-sync.json")
	type bundle struct {
		Favorites []string          `json:"favorites"`
		Presets   []presetData      `json:"presets"`
		Counts    map[string]int    `json:"play_counts"`
		Exported  time.Time         `json:"exported"`
	}
	var favs []string
	for p, ok := range s.favorites {
		if ok {
			favs = append(favs, p)
		}
	}
	b := bundle{Favorites: favs, Presets: s.presetsToSettings(), Counts: s.playCounts, Exported: time.Now()}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		s.status = "Sync export failed."
		invalidate()
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		s.status = "Sync export failed: " + err.Error()
	} else {
		s.status = "Exported sync bundle: " + path
	}
	invalidate()
}
