package appIcon

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"
)

var linuxIconSpecs = []iconSpec{
	{16, "16x16"},
	{22, "22x22"},
	{24, "24x24"},
	{32, "32x32"},
	{48, "48x48"},
	{64, "64x64"},
	{96, "96x96"},
	{128, "128x128"},
	{256, "256x256"},
	{512, "512x512"},
}

func generateLinux(src image.Image, outDir, appName string) error {
	zipPath := filepath.Join(outDir, "linux.zip")

	var entries []imageEntry
	for _, s := range linuxIconSpecs {
		entries = append(entries, imageEntry{
			Img:     resizeImg(src, s.Size),
			ZipPath: filepath.Join("hicolor", s.Name, "apps", appName+".png"),
		})
	}

	desktop := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s
Icon=%s
Terminal=false
`, appName, strings.ToLower(appName), appName)

	if err := createZip(zipPath, entries, nil, map[string]string{appName + ".desktop": desktop}); err != nil {
		return err
	}
	fmt.Printf("    → %s\n", zipPath)
	return nil
}
