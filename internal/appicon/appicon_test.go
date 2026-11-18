package appIcon

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func makeSquarePNG(size int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 128, 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func writeTestPNG(t *testing.T, size int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "src.png")
	os.WriteFile(path, makeSquarePNG(size), 0644)
	return path
}

func TestParsePlatforms(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"all", "all", []string{"ios", "android", "macos", "windows", "linux", "watchos", "web"}},
		{"单平台", "ios", []string{"ios"}},
		{"逗号分隔", "ios,android,macos", []string{"ios", "android", "macos"}},
		{"多余空格", "  IOS , Android  ", []string{"ios", "android"}},
		{"空元素跳过", "ios,,android,", []string{"ios", "android"}},
		{"混合大小写", "iOS,ANDROID", []string{"ios", "android"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePlatforms(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePlatforms(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDedupEntries(t *testing.T) {
	entries := []iosEntry{
		{FileName: "a.png"},
		{FileName: "b.png"},
		{FileName: "a.png"},
		{FileName: "c.png"},
		{FileName: "b.png"},
	}
	got := dedupEntries(entries)
	if len(got) != 3 {
		t.Fatalf("去重后应该3个, 实际 %d: %v", len(got), got)
	}
	if got[0].FileName != "a.png" || got[1].FileName != "b.png" || got[2].FileName != "c.png" {
		t.Errorf("顺序应该保持首次出现: %v", got)
	}
}

func TestDedupEntries_AllUnique(t *testing.T) {
	entries := []iosEntry{{FileName: "a"}, {FileName: "b"}, {FileName: "c"}}
	got := dedupEntries(entries)
	if len(got) != 3 {
		t.Errorf("全部唯一不应有变化: %v", got)
	}
}

func TestBuildIOSContents(t *testing.T) {
	entries := []iosEntry{
		{PixelSize: 100, FileName: "a.png", Size: "100x100", Idiom: "iphone", Scale: "1x", Role: "", Subtype: ""},
		{PixelSize: 200, FileName: "b.png", Size: "200x200", Idiom: "ipad", Scale: "2x", Role: "appLauncher", Subtype: "42mm"},
	}
	c := buildIOSContents(entries)

	if c.Info.Author != "ql-cli" || c.Info.Version != 1 {
		t.Errorf("Info 不对: %+v", c.Info)
	}
	if len(c.Images) != 2 {
		t.Fatalf("Images 数量不对: %d", len(c.Images))
	}
	if c.Images[1].Role != "appLauncher" || c.Images[1].Subtype != "42mm" {
		t.Errorf("可选字段丢失: %+v", c.Images[1])
	}

	_, err := json.Marshal(c)
	if err != nil {
		t.Errorf("应该能序列化成 JSON: %v", err)
	}
}

func TestLoadImage(t *testing.T) {
	t.Run("有效 PNG", func(t *testing.T) {
		path := writeTestPNG(t, 1024)
		img, err := loadImage(path)
		if err != nil {
			t.Fatal(err)
		}
		b := img.Bounds()
		if b.Dx() != 1024 || b.Dy() != 1024 {
			t.Errorf("尺寸不对: %dx%d", b.Dx(), b.Dy())
		}
	})

	t.Run("文件不存在", func(t *testing.T) {
		_, err := loadImage("/nonexistent.png")
		if err == nil {
			t.Fatal("应该报错")
		}
	})

	t.Run("非图片文件", func(t *testing.T) {
		bad := filepath.Join(t.TempDir(), "bad.png")
		os.WriteFile(bad, []byte("not a png"), 0644)
		_, err := loadImage(bad)
		if err == nil {
			t.Fatal("应该报错")
		}
	})
}

func TestResizeImg(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 500, 500))
	got := resizeImg(src, 100)
	b := got.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Errorf("缩放后尺寸应为 100x100, 实际 %dx%d", b.Dx(), b.Dy())
	}
}

func TestEncodePNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	data := encodePNG(img)
	if len(data) == 0 {
		t.Fatal("应该编码出 PNG 数据")
	}
	if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		t.Error("PNG 魔数不对")
	}
}

func TestCreateZip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "test.zip")

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	entries := []imageEntry{
		{Spec: iconSpec{Size: 10, Name: "a.png"}, Img: img, ZipPath: "icons/a.png"},
		{Spec: iconSpec{Size: 20, Name: "b.png"}, Img: img, ZipPath: "icons/b.png"},
	}
	extraFiles := map[string][]byte{
		"manifest.json": []byte(`{"name":"test"}`),
	}
	extraText := map[string]string{
		"readme.txt": "hello",
	}

	if err := createZip(zipPath, entries, extraFiles, extraText); err != nil {
		t.Fatal(err)
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	names := make(map[string]bool)
	for _, f := range zr.File {
		names[f.Name] = true
	}
	for _, want := range []string{"icons/a.png", "icons/b.png", "manifest.json", "readme.txt"} {
		if !names[want] {
			t.Errorf("zip 中缺条目: %s", want)
		}
	}
}

func TestGenerate_SmallSquare(t *testing.T) {
	dir := t.TempDir()
	src := writeTestPNG(t, 512)

	err := Generate(Options{
		SourcePath: src,
		OutDir:     dir,
		Platforms:  []string{"ios"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "ios.zip")); err != nil {
		t.Errorf("应该生成 ios.zip: %v", err)
	}
}

func TestGenerate_NotSquare(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rect.png")
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	var buf bytes.Buffer
	png.Encode(&buf, img)
	os.WriteFile(src, buf.Bytes(), 0644)

	err := Generate(Options{
		SourcePath: src,
		OutDir:     filepath.Join(dir, "out"),
		Platforms:  []string{"ios"},
	})
	if err == nil {
		t.Fatal("非正方形图片应该报错")
	}
}

func TestGenerate_UnknownPlatform(t *testing.T) {
	dir := t.TempDir()
	src := writeTestPNG(t, 1024)

	err := Generate(Options{
		SourcePath: src,
		OutDir:     dir,
		Platforms:  []string{"fantasy-os"},
	})
	if err != nil {
		t.Errorf("单个未知平台 Generate 整体应该不 return error (只是打印失败), 实际 error: %v", err)
	}

	err = generatePlatform("fantasy-os", image.NewRGBA(image.Rect(0, 0, 100, 100)), dir, "app")
	if err == nil {
		t.Error("generatePlatform 对未知平台必须返回 error")
	}
}

func TestGenerate_AllPlatforms(t *testing.T) {
	dir := t.TempDir()
	src := writeTestPNG(t, 1024)

	err := Generate(Options{
		SourcePath: src,
		OutDir:     dir,
		AppName:    "MyApp",
	})
	if err != nil {
		t.Fatalf("全部平台生成失败: %v", err)
	}

	expected := []string{"ios.zip", "android.zip", "macos.zip", "windows.zip", "linux.zip", "watchos.zip", "web.zip"}
	for _, name := range expected {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("缺产物: %s (%v)", name, err)
		}
	}
}

func TestGeneratePlatform_SwitchCoverage(t *testing.T) {
	dir := t.TempDir()
	src := image.NewRGBA(image.Rect(0, 0, 1024, 1024))

	for _, p := range []string{"ios", "watchos", "android", "macos", "windows", "linux", "web"} {
		t.Run(p, func(t *testing.T) {
			err := generatePlatform(p, src, dir, "TestApp")
			if err != nil {
				t.Errorf("generatePlatform(%s) error: %v", p, err)
			}
		})
	}

	t.Run("unknown", func(t *testing.T) {
		err := generatePlatform("nosuch", src, dir, "TestApp")
		if err == nil {
			t.Error("未知平台应该报错")
		}
	})
}
