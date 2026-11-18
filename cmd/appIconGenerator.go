package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var iconGenCmd = &cobra.Command{
	Use:   "icongen",
	Short: color.CyanString("应用图标生成器 - 全平台图标输出"),
	Long: color.CyanString(`从一张高清图片源(PNG/JPG/JPEG/GIF/BMP/WEBP)自动生成全平台应用图标:
  iOS, watchOS   → AppIcon.appiconset + Contents.json
  Android        → mipmap-*/ 目录 + adaptive-icon
  macOS          → .icns 文件
  Windows        → .ico 文件
  Linux          → hicolor/ 目录 + .desktop
  Web PWA        → manifest.json + PNG icons
并按平台打包为 ZIP 文件。`),
	RunE: runIconGen,
}

var (
	iconSource  string
	iconOutput  string
	iconName    string
	iconFormats string
)

func init() {
	iconGenCmd.Flags().StringVarP(&iconSource, "source", "s", "", "源图片路径(支持PNG/JPG/JPEG/BMP/WEBP, 建议1024x1024)")
	iconGenCmd.Flags().StringVarP(&iconOutput, "dir", "d", "./icons", "输出目录")
	iconGenCmd.Flags().StringVarP(&iconName, "name", "n", "AppIcon", "图标/应用名称")
	iconGenCmd.Flags().StringVarP(&iconFormats, "formats", "f", "all", "平台列表,逗号分隔: ios,android,macos,windows,linux,watchos,web (默认all)")
	rootCmd.AddCommand(iconGenCmd)
}

type iconSpec struct {
	Size int
	Name string
}

type imageEntry struct {
	Spec    iconSpec
	Img     image.Image
	ZipPath string
}

func runIconGen(cmd *cobra.Command, args []string) error {
	if iconSource == "" {
		return fmt.Errorf("请通过 -s 指定源PNG图标路径")
	}
	if _, err := os.Stat(iconSource); err != nil {
		return fmt.Errorf("源图标不存在: %w", err)
	}

	srcImg, err := loadImage(iconSource)
	if err != nil {
		return fmt.Errorf("加载源图标失败: %w", err)
	}
	b := srcImg.Bounds()
	if b.Dx() != b.Dy() {
		return fmt.Errorf("源图标必须是正方形, 当前 %dx%d", b.Dx(), b.Dy())
	}
	if b.Dx() < 1024 {
		fmt.Printf("⚠ 源图 %dx%d 小于建议的 1024x1024, 部分平台高清图标可能模糊\n", b.Dx(), b.Dy())
	}

	if err := os.MkdirAll(iconOutput, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	platforms := parsePlatforms(iconFormats)

	fmt.Printf("源图: %s (%dx%d)\n输出: %s\n平台: %v\n\n", iconSource, b.Dx(), b.Dy(), iconOutput, platforms)

	for _, p := range platforms {
		fmt.Printf("▸ 生成 %s ...\n", p)
		if err := generatePlatform(p, srcImg, iconOutput, iconName); err != nil {
			fmt.Printf("  ✗ %s 失败: %v\n", p, err)
		} else {
			fmt.Printf("  ✓ %s 完成\n", p)
		}
	}
	return nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func parsePlatforms(s string) []string {
	if s == "all" {
		return []string{"ios", "android", "macos", "windows", "linux", "watchos", "web"}
	}
	raw := strings.Split(s, ",")
	var out []string
	for _, r := range raw {
		r = strings.TrimSpace(strings.ToLower(r))
		if r != "" {
			out = append(out, r)
		}
	}
	return out
}

// ============================================================
//  iOS / watchOS
// ============================================================

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

// ============================================================
//  Android
// ============================================================

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

// ============================================================
//  macOS ICNS
// ============================================================

var macosIconSpecs = []iconSpec{
	{16, "icon_16x16"},
	{32, "icon_16x16@2x"},
	{32, "icon_32x32"},
	{64, "icon_32x32@2x"},
	{128, "icon_128x128"},
	{256, "icon_128x128@2x"},
	{256, "icon_256x256"},
	{512, "icon_256x256@2x"},
	{512, "icon_512x512"},
	{1024, "icon_512x512@2x"},
}

type icnsEntry struct {
	OSType string
	Size   int
	PNG    []byte
}

var icnsTypeMap = map[int]string{
	16:   "icp4",
	32:   "icp5",
	64:   "icp6",
	128:  "icp7",
	256:  "icp8",
	512:  "icp9",
	1024: "ic10",
}

func encodeICNS(images []icnsEntry) ([]byte, error) {
	var body bytes.Buffer
	for _, img := range images {
		if imgSize, ok := icnsTypeMap[img.Size]; ok {
			body.WriteString(imgSize)
			entryLen := 8 + len(img.PNG)
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, uint32(entryLen))
			body.Write(buf)
			body.Write(img.PNG)
		}
	}

	var out bytes.Buffer
	out.WriteString("icns")
	totalLen := 8 + body.Len()
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(totalLen))
	out.Write(buf)
	out.Write(body.Bytes())
	return out.Bytes(), nil
}

// ============================================================
//  Windows ICO
// ============================================================

func encodeICO(images []struct {
	Size int
	PNG  []byte
}) ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.LittleEndian, uint16(0))
	binary.Write(&out, binary.LittleEndian, uint16(1))
	binary.Write(&out, binary.LittleEndian, uint16(len(images)))

	dataOffset := 6 + len(images)*16
	for _, img := range images {
		w := img.Size
		if w >= 256 {
			w = 0
		}
		h := img.Size
		if h >= 256 {
			h = 0
		}
		out.WriteByte(byte(w))
		out.WriteByte(byte(h))
		out.WriteByte(0)
		out.WriteByte(0)
		binary.Write(&out, binary.LittleEndian, uint16(1))
		binary.Write(&out, binary.LittleEndian, uint16(32))
		binary.Write(&out, binary.LittleEndian, uint32(len(img.PNG)))
		binary.Write(&out, binary.LittleEndian, uint32(dataOffset))
		dataOffset += len(img.PNG)
	}
	for _, img := range images {
		out.Write(img.PNG)
	}
	return out.Bytes(), nil
}

// ============================================================
//  Linux - hicolor
// ============================================================

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

// ============================================================
//  Web PWA
// ============================================================

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

// ============================================================
//  核心生成逻辑
// ============================================================

func resizeImg(src image.Image, size int) image.Image {
	return imaging.Resize(src, size, size, imaging.Lanczos)
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func generatePlatform(platform string, src image.Image, outDir string, appName string) error {

	entries := []imageEntry{}
	extraFiles := map[string][]byte{}
	extraText := map[string]string{}

	switch platform {
	case "ios":
		for _, e := range dedupEntries(iosEntries) {
			entries = append(entries, imageEntry{
				Img:     resizeImg(src, e.PixelSize),
				ZipPath: "AppIcon.appiconset/" + e.FileName,
			})
		}
		c := buildIOSContents(iosEntries)
		b, _ := json.MarshalIndent(c, "", "  ")
		extraFiles["AppIcon.appiconset/Contents.json"] = b

	case "watchos":
		for _, e := range dedupEntries(watchOSEntries) {
			entries = append(entries, imageEntry{
				Img:     resizeImg(src, e.PixelSize),
				ZipPath: "AppIcon.appiconset/" + e.FileName,
			})
		}
		c := buildIOSContents(watchOSEntries)
		b, _ := json.MarshalIndent(c, "", "  ")
		extraFiles["AppIcon.appiconset/Contents.json"] = b

	case "android":
		for _, m := range androidMipmaps {
			entries = append(entries, imageEntry{
				Spec:    iconSpec{m.Size, "ic_launcher.png"},
				Img:     resizeImg(src, m.Size),
				ZipPath: fmt.Sprintf("mipmap-anydpi-v26/%s/%s", m.Dir, "ic_launcher.png"),
			})
			entries = append(entries, imageEntry{
				Spec:    iconSpec{m.Size, "ic_launcher_round.png"},
				Img:     resizeImg(src, m.Size),
				ZipPath: fmt.Sprintf("mipmap-anydpi-v26/%s/%s", m.Dir, "ic_launcher_round.png"),
			})
			entries = append(entries, imageEntry{
				Spec:    iconSpec{m.Size, "ic_launcher.png"},
				Img:     resizeImg(src, m.Size),
				ZipPath: fmt.Sprintf("%s/%s", m.Dir, "ic_launcher.png"),
			})
			entries = append(entries, imageEntry{
				Spec:    iconSpec{m.Size, "ic_launcher_round.png"},
				Img:     resizeImg(src, m.Size),
				ZipPath: fmt.Sprintf("%s/%s", m.Dir, "ic_launcher_round.png"),
			})
		}
		adaptiveFg := resizeImg(src, 108)
		entries = append(entries, imageEntry{
			Spec:    iconSpec{108, "ic_launcher_foreground.png"},
			Img:     adaptiveFg,
			ZipPath: "mipmap-anydpi-v26/ic_launcher_foreground.png",
		})
		entries = append(entries, imageEntry{
			Spec:    iconSpec{108, "ic_launcher_foreground.png"},
			Img:     adaptiveFg,
			ZipPath: "mipmap-anydpi-v26/ic_launcher_round_foreground.png",
		})

	case "macos":
		var icnsImgs []icnsEntry
		for _, s := range macosIconSpecs {
			img := resizeImg(src, s.Size)
			pngData := encodePNG(img)
			icnsImgs = append(icnsImgs, icnsEntry{Size: s.Size, PNG: pngData})
		}
		sort.Slice(icnsImgs, func(i, j int) bool { return icnsImgs[i].Size < icnsImgs[j].Size })
		icnsData, err := encodeICNS(icnsImgs)
		if err != nil {
			return err
		}
		extraFiles[appName+".icns"] = icnsData

	case "windows":
		var icoImgs []struct {
			Size int
			PNG  []byte
		}
		for _, s := range []int{16, 24, 32, 48, 64, 128, 256} {
			icoImgs = append(icoImgs, struct {
				Size int
				PNG  []byte
			}{s, encodePNG(resizeImg(src, s))})
		}
		icoData, err := encodeICO(icoImgs)
		if err != nil {
			return err
		}
		extraFiles[appName+".ico"] = icoData

	case "linux":
		for _, s := range linuxIconSpecs {
			entries = append(entries, imageEntry{
				Spec:    iconSpec{s.Size, appName + ".png"},
				Img:     resizeImg(src, s.Size),
				ZipPath: filepath.Join("hicolor", s.Name, "apps", appName+".png"),
			})
		}
		extraText[appName+".desktop"] = fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s
Icon=%s
Terminal=false
`, appName, strings.ToLower(appName), appName)

	case "web":
		for _, s := range webIconSpecs {
			entries = append(entries, imageEntry{
				Spec:    s,
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
		extraFiles["manifest.json"] = b
		extraText["browserconfig.xml"] = fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<browserconfig>
  <msapplication>
    <tile>
      <square150x150logo src="mstile-150x150.png"/>
    </tile>
  </msapplication>
</browserconfig>
`)

	default:
		return fmt.Errorf("未知平台: %s", platform)
	}

	zipPath := filepath.Join(outDir, platform+".zip")
	if err := createPlatformZip(zipPath, entries, extraFiles, extraText); err != nil {
		return err
	}
	fmt.Printf("    → %s\n", zipPath)
	return nil
}

func createPlatformZip(zipPath string, entries []imageEntry, extraFiles map[string][]byte, extraText map[string]string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, e := range entries {
		w, err := zw.Create(e.ZipPath)
		if err != nil {
			return err
		}
		img := encodePNG(e.Img)
		if _, err := w.Write(img); err != nil {
			return err
		}
	}

	for name, data := range extraFiles {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	for name, text := range extraText {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, text); err != nil {
			return err
		}
	}
	return nil
}
