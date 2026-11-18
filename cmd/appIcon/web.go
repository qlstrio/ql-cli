package appIcon

import (
	"encoding/json"
	"fmt"
	"image"
	"path/filepath"
)

var webIconSpecs = []iconSpec{
	{192, "icon-192.png"},
	{512, "icon-512.png"},
	{180, "apple-touch-icon.png"},
	{150, "mstile-150x150.png"},
}

type pwaManifest struct {
	Name       string    `json:"name"`
	Short      string    `json:"short_name"`
	Icons      []pwaIcon `json:"icons"`
	StartURL   string    `json:"start_url"`
	Display    string    `json:"display"`
	Background string    `json:"background_color"`
	Theme      string    `json:"theme_color"`
}

type pwaIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

func generateWeb(src image.Image, outDir, appName string) error {
	zipPath := filepath.Join(outDir, "web.zip")

	var entries []imageEntry
	for _, s := range webIconSpecs {
		entries = append(entries, imageEntry{
			Img:     resizeImg(src, s.Size),
			ZipPath: s.Name,
		})
	}

	m := pwaManifest{
		Name:       appName,
		Short:      appName,
		StartURL:   "/",
		Display:    "standalone",
		Background: "#ffffff",
		Theme:      "#000000",
		Icons: []pwaIcon{
			{"icon-192.png", "192x192", "image/png", "any maskable"},
			{"icon-512.png", "512x512", "image/png", "any maskable"},
		},
	}
	b, _ := json.MarshalIndent(m, "", "  ")

	browserconfig := `<?xml version="1.0" encoding="utf-8"?>
<browserconfig>
  <msapplication>
    <tile>
      <square150x150logo src="mstile-150x150.png"/>
    </tile>
  </msapplication>
</browserconfig>
`

	if err := createZip(zipPath, entries,
		map[string][]byte{"manifest.json": b},
		map[string]string{"browserconfig.xml": browserconfig},
	); err != nil {
		return err
	}
	fmt.Printf("    → %s\n", zipPath)
	return nil
}
