package appIcon

import (
	"fmt"
	"image"
	"path/filepath"
)

var androidMipmaps = []struct {
	Dir  string
	Size int
}{
	{"mipmap-mdpi", 48},
	{"mipmap-hdpi", 72},
	{"mipmap-xhdpi", 96},
	{"mipmap-xxhdpi", 144},
	{"mipmap-xxxhdpi", 192},
}

func generateAndroid(src image.Image, outDir string) error {
	zipPath := filepath.Join(outDir, "android.zip")

	var entries []imageEntry
	for _, m := range androidMipmaps {
		entries = append(entries, imageEntry{
			Img:     resizeImg(src, m.Size),
			ZipPath: fmt.Sprintf("%s/ic_launcher.png", m.Dir),
		})
		entries = append(entries, imageEntry{
			Img:     resizeImg(src, m.Size),
			ZipPath: fmt.Sprintf("%s/ic_launcher_round.png", m.Dir),
		})
	}

	adaptiveFg := resizeImg(src, 108)
	entries = append(entries,
		imageEntry{Img: adaptiveFg, ZipPath: "mipmap-anydpi-v26/ic_launcher_foreground.png"},
		imageEntry{Img: adaptiveFg, ZipPath: "mipmap-anydpi-v26/ic_launcher_round_foreground.png"},
	)

	if err := createZip(zipPath, entries, nil, nil); err != nil {
		return err
	}
	fmt.Printf("    → %s\n", zipPath)
	return nil
}
