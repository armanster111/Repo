package metadata

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

// TrackInfo contains display metadata for an audio file.
type TrackInfo struct {
	Title  string
	Artist string
	Album  string
	Genre  string
}

// ArtGrid is a low-resolution color grid for blurred album-art backgrounds.
type ArtGrid struct {
	Cols   int
	Rows   int
	Colors []uint32 // 0x00RRGGBB
}

// Read loads tags and album art from a file.
func Read(path string) (TrackInfo, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return TrackInfo{}, nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return TrackInfo{Title: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))}, nil, err
	}

	info := TrackInfo{
		Title:  firstNonEmpty(m.Title(), strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))),
		Artist: m.Artist(),
		Album:  m.Album(),
		Genre:  m.Genre(),
	}
	picture := m.Picture()
	if picture == nil {
		return info, nil, nil
	}
	return info, picture.Data, nil
}

// ArtGridFromBytes downsamples cover art into a color grid.
func ArtGridFromBytes(data []byte, cols, rows int) *ArtGrid {
	if len(data) == 0 || cols <= 0 || rows <= 0 {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	bounds := img.Bounds()
	grid := &ArtGrid{Cols: cols, Rows: rows, Colors: make([]uint32, cols*rows)}
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			sx := bounds.Min.X + (x*bounds.Dx())/cols
			sy := bounds.Min.Y + (y*bounds.Dy())/rows
			r, g, b, _ := img.At(sx, sy).RGBA()
			grid.Colors[y*cols+x] = uint32((r>>8)<<16) | uint32((g>>8)<<8) | uint32(b>>8)
		}
	}
	return grid
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
