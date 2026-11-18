package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func getAbsDirPath(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf(color.RedString("获取目录绝对路径失败: %w"), err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return "", fmt.Errorf(color.RedString("目录不存在: %w"), err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf(color.RedString("路径不是目录: %s"), absDir)
	}
	return absDir, nil
}
