//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armanster111/muse/internal/library"
)

func (s *appState) defaultLibraryRoots() []string {
	if len(s.libraryPaths) > 0 {
		return s.libraryPaths
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "Music"),
		filepath.Join(home, "Downloads"),
	}
	if s.libraryRoot != "" {
		candidates = append([]string{s.libraryRoot}, candidates...)
	}
	var roots []string
	seen := map[string]bool{}
	for _, root := range candidates {
		root = filepath.Clean(root)
		if root == "" || seen[root] {
			continue
		}
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			roots = append(roots, root)
			seen[root] = true
		}
	}
	return roots
}

func (s *appState) scanLibraryAsync() {
	if s.libraryScanning {
		return
	}
	roots := s.defaultLibraryRoots()
	if len(roots) == 0 {
		return
	}
	s.libraryScanning = true
	s.status = "Scanning music library..."
	invalidate()
	go func() {
		var merged []library.Entry
		for _, root := range roots {
			idx, err := library.Scan(root)
			if err != nil || idx == nil {
				continue
			}
			merged = append(merged, idx.Entries...)
		}
		combined := &library.Index{
			Root:     strings.Join(roots, ";"),
			Entries:  merged,
			ByArtist: map[string][]library.Entry{},
			ByAlbum:  map[string][]library.Entry{},
			ByGenre:  map[string][]library.Entry{},
		}
		combined.ApplyPlayCounts(s.playCounts, s.lastPlayedTimes())
		s.libraryIndex = combined
		s.libraryScanning = false
		s.updateStatus()
		invalidate()
	}()
}

func (s *appState) lastPlayedTimes() map[string]time.Time {
	out := make(map[string]time.Time, len(s.lastPlayed))
	for path, unix := range s.lastPlayed {
		out[path] = time.Unix(unix, 0)
	}
	return out
}

func (s *appState) showLibraryView() {
	s.panel = viewLibrary
	s.panelGroup = ""
	s.panelGroupKey = ""
	s.panelRow = 0
	invalidate()
}

func (s *appState) showAlbumsView() {
	s.panel = viewAlbums
	s.panelGroup = ""
	s.panelGroupKey = ""
	s.panelRow = 0
	invalidate()
}

func (s *appState) showArtistsView() {
	s.panel = viewArtists
	s.panelGroup = ""
	s.panelGroupKey = ""
	s.panelRow = 0
	invalidate()
}

func (s *appState) showSmartView() {
	s.panel = viewSmart
	s.panelRow = 0
	invalidate()
}

func (s *appState) libraryEntriesForPanel() []library.Entry {
	if s.libraryIndex == nil {
		return nil
	}
	switch s.panel {
	case viewArtists:
		if s.panelGroupKey == "" {
			return nil
		}
		return s.libraryIndex.ByArtist[s.panelGroupKey]
	case viewAlbums:
		if s.panelGroupKey == "" {
			return nil
		}
		return s.libraryIndex.ByAlbum[s.panelGroupKey]
	case viewSmart:
		return s.libraryIndex.MostPlayed(64)
	default:
		return s.libraryIndex.Entries
	}
}

func (s *appState) libraryGroupsForPanel() []string {
	if s.libraryIndex == nil {
		return nil
	}
	switch s.panel {
	case viewArtists:
		return s.libraryIndex.ArtistNames()
	case viewAlbums:
		return s.libraryIndex.AlbumKeys()
	default:
		return nil
	}
}

func (s *appState) openLibraryEntry(entry library.Entry) {
	if entry.Path == "" {
		return
	}
	s.openPath(entry.Path)
}

func (s *appState) bumpPlayCount(path string) {
	clean := filepath.Clean(path)
	s.playCounts[clean]++
	s.lastPlayed[clean] = time.Now().Unix()
	if s.libraryIndex != nil {
		s.libraryIndex.ApplyPlayCounts(s.playCounts, s.lastPlayedTimes())
	}
	s.saveSettings()
}

func (s *appState) libraryPanelTitle() string {
	switch s.panel {
	case viewArtists:
		if s.panelGroupKey != "" {
			return "Artist: " + truncate(s.panelGroupKey, 24)
		}
		return "Artists"
	case viewAlbums:
		if s.panelGroupKey != "" {
			return truncate(s.panelGroupKey, 28)
		}
		return "Albums"
	case viewSmart:
		return "Top Played"
	default:
		return "Library"
	}
}

func (s *appState) cycleLibraryPanel() {
	switch s.panel {
	case viewQueue, viewRecent, viewFavorites, viewSearch:
		s.showLibraryView()
	case viewLibrary:
		s.showArtistsView()
	case viewArtists:
		s.showAlbumsView()
	case viewAlbums:
		s.showSmartView()
	default:
		s.panel = viewQueue
	}
	invalidate()
}

func libraryStatusText(scanning bool, count int) string {
	if scanning {
		return "Scanning library..."
	}
	return fmt.Sprintf("Library: %d tracks", count)
}
