package library

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/armanster111/muse/internal/metadata"
)

var audioExtensions = map[string]bool{
	".wav": true, ".mp3": true, ".wma": true, ".mid": true, ".midi": true,
	".aiff": true, ".aif": true, ".au": true, ".snd": true, ".flac": true, ".ogg": true,
}

// Entry is one indexed track in the music library.
type Entry struct {
	Path         string    `json:"path"`
	Title        string    `json:"title"`
	Artist       string    `json:"artist"`
	Album        string    `json:"album"`
	Genre        string    `json:"genre"`
	PlayCount    int       `json:"play_count"`
	LastPlayed   time.Time `json:"last_played"`
	ModifiedTime time.Time `json:"modified_time"`
}

// Index is a scanned music library with grouped views.
type Index struct {
	Entries   []Entry            `json:"entries"`
	ByArtist  map[string][]Entry `json:"-"`
	ByAlbum   map[string][]Entry `json:"-"`
	ByGenre   map[string][]Entry `json:"-"`
	ScannedAt time.Time          `json:"scanned_at"`
	Root      string             `json:"root"`
}

// Scan walks a directory tree and indexes supported audio files.
func Scan(root string) (*Index, error) {
	idx := &Index{
		Root:     filepath.Clean(root),
		ByArtist: map[string][]Entry{},
		ByAlbum:  map[string][]Entry{},
		ByGenre:  map[string][]Entry{},
	}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := strings.ToLower(d.Name())
			if base == "$recycle.bin" || base == "system volume information" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isAudio(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		entry := Entry{
			Path:         filepath.Clean(path),
			ModifiedTime: info.ModTime(),
		}
		if meta, _, err := metadata.Read(path); err == nil {
			entry.Title = meta.Title
			entry.Artist = meta.Artist
			entry.Album = meta.Album
			entry.Genre = meta.Genre
		}
		if entry.Title == "" {
			entry.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
		if entry.Artist == "" {
			entry.Artist = "Unknown Artist"
		}
		if entry.Album == "" {
			entry.Album = "Unknown Album"
		}
		if entry.Genre == "" {
			entry.Genre = "Unknown"
		}
		idx.Entries = append(idx.Entries, entry)
		return nil
	})
	if err != nil {
		return nil, err
	}
	idx.rebuildGroups()
	idx.ScannedAt = time.Now()
	return idx, nil
}

func (idx *Index) rebuildGroups() {
	idx.ByArtist = map[string][]Entry{}
	idx.ByAlbum = map[string][]Entry{}
	idx.ByGenre = map[string][]Entry{}
	for _, e := range idx.Entries {
		idx.ByArtist[e.Artist] = append(idx.ByArtist[e.Artist], e)
		key := e.Artist + " — " + e.Album
		idx.ByAlbum[key] = append(idx.ByAlbum[key], e)
		idx.ByGenre[e.Genre] = append(idx.ByGenre[e.Genre], e)
	}
	for k := range idx.ByArtist {
		sort.Slice(idx.ByArtist[k], func(i, j int) bool {
			return strings.ToLower(idx.ByArtist[k][i].Title) < strings.ToLower(idx.ByArtist[k][j].Title)
		})
	}
	sort.Slice(idx.Entries, func(i, j int) bool {
		return strings.ToLower(idx.Entries[i].Title) < strings.ToLower(idx.Entries[j].Title)
	})
}

// ApplyPlayCounts merges persisted play statistics into the index.
func (idx *Index) ApplyPlayCounts(counts map[string]int, last map[string]time.Time) {
	for i := range idx.Entries {
		path := idx.Entries[i].Path
		idx.Entries[i].PlayCount = counts[path]
		idx.Entries[i].LastPlayed = last[path]
	}
	idx.rebuildGroups()
}

// MostPlayed returns entries sorted by play count.
func (idx *Index) MostPlayed(limit int) []Entry {
	out := append([]Entry(nil), idx.Entries...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].PlayCount == out[j].PlayCount {
			return out[i].LastPlayed.After(out[j].LastPlayed)
		}
		return out[i].PlayCount > out[j].PlayCount
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// RecentlyAdded returns newest files by modified time.
func (idx *Index) RecentlyAdded(limit int) []Entry {
	out := append([]Entry(nil), idx.Entries...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ModifiedTime.After(out[j].ModifiedTime)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// ArtistNames returns sorted unique artist names.
func (idx *Index) ArtistNames() []string {
	names := make([]string, 0, len(idx.ByArtist))
	for name := range idx.ByArtist {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// AlbumKeys returns sorted album keys (Artist — Album).
func (idx *Index) AlbumKeys() []string {
	keys := make([]string, 0, len(idx.ByAlbum))
	for key := range idx.ByAlbum {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func isAudio(path string) bool {
	return audioExtensions[strings.ToLower(filepath.Ext(path))]
}
