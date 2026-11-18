package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/qlstrio/ql-cli/cmd/appIcon"
)

var iconGenCmd = &cobra.Command{
	Use:   "icongen",
	Short: color.CyanString("应用图标生成器 - 全平台图标输出"),
	Long: color.CyanString(`从一张高清图片源(PNG/JPG/JPEG/BMP/WEBP)自动生成全平台应用图标:
  iOS, watchOS   → AppIcon.appiconset + Contents.json
  Android        → mipmap-*/ 目录
  macOS          → .icns 文件
  Windows        → .ico 文件
  Linux          → hicolor/ 目录 + .desktop
  Web PWA        → manifest.json + browserconfig.xml
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

func runIconGen(cmd *cobra.Command, args []string) error {
	if iconSource == "" {
		return fmt.Errorf(color.RedString("请通过 -s 指定源图片路径"))
	}
	if _, err := os.Stat(iconSource); err != nil {
		return fmt.Errorf(color.RedString("源图标不存在: %w"), err)
	}

	if err := os.MkdirAll(iconOutput, 0755); err != nil {
		return fmt.Errorf(color.RedString("创建输出目录失败: %w"), err)
	}

	var platforms []string
	if iconFormats != "all" {
		for _, r := range strings.Split(iconFormats, ",") {
			r = strings.TrimSpace(strings.ToLower(r))
			if r != "" {
				platforms = append(platforms, r)
			}
		}
	}

	return appIcon.Generate(appIcon.Options{
		SourcePath: iconSource,
		OutDir:     iconOutput,
		AppName:    iconName,
		Platforms:  platforms,
	})
}
