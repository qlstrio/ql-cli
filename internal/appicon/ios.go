package appIcon

import (
	"encoding/json"
	"fmt"
	"image"
	"path/filepath"
)

type iosEntry struct {
	PixelSize int
	FileName  string
	Size      string
	Idiom     string
	Scale     string
	Role      string
	Subtype   string
}

var iosEntries = []iosEntry{
	{20, "Icon-App-20x20@1x.png", "20x20", "iphone", "1x", "", ""},
	{40, "Icon-App-20x20@2x.png", "20x20", "iphone", "2x", "", ""},
	{60, "Icon-App-20x20@3x.png", "20x20", "iphone", "3x", "", ""},
	{29, "Icon-App-29x29@1x.png", "29x29", "iphone", "1x", "", ""},
	{58, "Icon-App-29x29@2x.png", "29x29", "iphone", "2x", "", ""},
	{87, "Icon-App-29x29@3x.png", "29x29", "iphone", "3x", "", ""},
	{40, "Icon-App-40x40@1x.png", "40x40", "iphone", "1x", "", ""},
	{80, "Icon-App-40x40@2x.png", "40x40", "iphone", "2x", "", ""},
	{120, "Icon-App-40x40@3x.png", "40x40", "iphone", "3x", "", ""},
	{120, "Icon-App-60x60@2x.png", "60x60", "iphone", "2x", "", ""},
	{180, "Icon-App-60x60@3x.png", "60x60", "iphone", "3x", "", ""},
	{40, "Icon-App-20x20@2x.png", "20x20", "ipad", "2x", "", ""},
	{60, "Icon-App-20x20@3x.png", "20x20", "ipad", "3x", "", ""},
	{29, "Icon-App-29x29@1x.png", "29x29", "ipad", "1x", "", ""},
	{58, "Icon-App-29x29@2x.png", "29x29", "ipad", "2x", "", ""},
	{76, "Icon-App-76x76@1x.png", "76x76", "ipad", "1x", "", ""},
	{152, "Icon-App-76x76@2x.png", "76x76", "ipad", "2x", "", ""},
	{167, "Icon-App-83.5x83.5@2x.png", "83.5x83.5", "ipad", "2x", "", ""},
	{1024, "ItunesArtwork@2x.png", "1024x1024", "ios-marketing", "1x", "", ""},
}

var watchOSEntries = []iosEntry{
	{48, "Icon-App-24x24@2x.png", "24x24", "watch", "2x", "notificationCenter", "38mm"},
	{55, "Icon-App-27.5x27.5@2x.png", "27.5x27.5", "watch", "2x", "notificationCenter", "42mm"},
	{58, "Icon-App-29x29@2x.png", "29x29", "watch", "2x", "companionSettings", ""},
	{87, "Icon-App-29x29@3x.png", "29x29", "watch", "3x", "companionSettings", ""},
	{80, "Icon-App-40x40@2x.png", "40x40", "watch", "2x", "appLauncher", "38mm"},
	{172, "Icon-App-86x86@2x.png", "86x86", "watch", "2x", "quickLook", "38mm"},
	{196, "Icon-App-98x98@2x.png", "98x98", "watch", "2x", "quickLook", "42mm"},
	{216, "Icon-App-108x108@2x.png", "108x108", "watch", "2x", "appLauncher", "44mm"},
}

type iosContents struct {
	Images []iosImageEntry `json:"images"`
	Info   iosInfo         `json:"info"`
}

type iosImageEntry struct {
	Size     string `json:"size"`
	Idiom    string `json:"idiom"`
	Filename string `json:"filename"`
	Scale    string `json:"scale"`
	Role     string `json:"role,omitempty"`
	Subtype  string `json:"subtype,omitempty"`
}

type iosInfo struct {
	Author  string `json:"author"`
	Version int    `json:"version"`
}

func dedupEntries(entries []iosEntry) []iosEntry {
	seen := map[string]bool{}
	var out []iosEntry
	for _, e := range entries {
		if !seen[e.FileName] {
			seen[e.FileName] = true
			out = append(out, e)
		}
	}
	return out
}

func buildIOSContents(entries []iosEntry) iosContents {
	var images []iosImageEntry
	for _, e := range entries {
		images = append(images, iosImageEntry{e.Size, e.Idiom, e.FileName, e.Scale, e.Role, e.Subtype})
	}
	return iosContents{Images: images, Info: iosInfo{Author: "ql-cli", Version: 1}}
}

func generateIOS(src image.Image, outDir, platform string, entries []iosEntry) error {
	zipPath := filepath.Join(outDir, platform+".zip")

	var imageEntries []imageEntry
	for _, e := range dedupEntries(entries) {
		imageEntries = append(imageEntries, imageEntry{
			Img:     resizeImg(src, e.PixelSize),
			ZipPath: "AppIcon.appiconset/" + e.FileName,
		})
	}

	c := buildIOSContents(entries)
	b, _ := json.MarshalIndent(c, "", "  ")
	extraFiles := map[string][]byte{"AppIcon.appiconset/Contents.json": b}

	if err := createZip(zipPath, imageEntries, extraFiles, nil); err != nil {
		return err
	}
	fmt.Printf("    → %s\n", zipPath)
	return nil
}
