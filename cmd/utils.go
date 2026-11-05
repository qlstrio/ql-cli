package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func getAbsDirPath(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("获取目录绝对路径失败: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return "", fmt.Errorf("目录不存在: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("路径不是目录: %s", absDir)
	}
	return absDir, nil
}
