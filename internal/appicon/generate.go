package appIcon

import (
	"fmt"
	"image"
	"strings"

	"github.com/fatih/color"
)

type Options struct {
	SourcePath string
	OutDir     string
	AppName    string
	Platforms  []string
}

func Generate(opts Options) error {
	srcImg, err := loadImage(opts.SourcePath)
	if err != nil {
		return fmt.Errorf("加载源图标失败: %w", err)
	}
	b := srcImg.Bounds()
	if b.Dx() != b.Dy() {
		return fmt.Errorf("源图标必须是正方形, 当前 %dx%d", b.Dx(), b.Dy())
	}
	if b.Dx() < 1024 {
		fmt.Printf(color.YellowString("⚠ 源图 %dx%d 小于建议的 1024x1024, 部分平台高清图标可能模糊\n"), b.Dx(), b.Dy())
	}

	platforms := opts.Platforms
	if len(platforms) == 0 {
		platforms = []string{"ios", "android", "macos", "windows", "linux", "watchos", "web"}
	}

	fmt.Printf("源图: %s (%dx%d)\n输出: %s\n平台: %v\n\n", opts.SourcePath, b.Dx(), b.Dy(), opts.OutDir, platforms)

	for _, p := range platforms {
		fmt.Printf("▸ 生成 %s ...\n", p)
		if err := generatePlatform(p, srcImg, opts.OutDir, opts.AppName); err != nil {
			fmt.Printf("  ✗ %s 失败: %v\n", p, err)
		} else {
			fmt.Printf("  ✓ %s 完成\n", p)
		}
	}
	return nil
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

func generatePlatform(platform string, src image.Image, outDir, appName string) error {
	switch platform {
	case "ios":
		return generateIOS(src, outDir, platform, iosEntries)
	case "watchos":
		return generateIOS(src, outDir, platform, watchOSEntries)
	case "android":
		return generateAndroid(src, outDir)
	case "macos":
		return generateMacos(src, outDir, appName)
	case "windows":
		return generateWindows(src, outDir, appName)
	case "linux":
		return generateLinux(src, outDir, appName)
	case "web":
		return generateWeb(src, outDir, appName)
	default:
		return fmt.Errorf("未知平台: %s", platform)
	}
}
