package splitZip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func DoExtract(zipFiles []string, unzipOutput string) (err error) {
	totalFiles := 0
	for _, zipPath := range zipFiles {
		n, err := extractZip(zipPath, unzipOutput)
		if err != nil {
			fmt.Printf("解压 %s 失败: %v\n", filepath.Base(zipPath), err)
			continue
		}
		totalFiles += n
		fmt.Printf("✓ %s -> %d 个文件\n", filepath.Base(zipPath), n)
	}
	fmt.Printf("完成: %d 个zip, 共还原 %d 个文件到 %s\n", len(zipFiles), totalFiles, unzipOutput)
	return nil
}

func extractZip(zipPath string, destDir string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("打开zip失败: %w", err)
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		cleanName := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanName, "..") {
			return 0, fmt.Errorf("zip内包含非法路径: %s", f.Name)
		}

		fullPath := filepath.Join(destDir, cleanName)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return 0, fmt.Errorf("创建目录 %s 失败: %w", fullPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return 0, fmt.Errorf("创建父目录 %s 失败: %w", filepath.Dir(fullPath), err)
		}

		outFile, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return 0, fmt.Errorf("创建文件 %s 失败: %w", fullPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return 0, fmt.Errorf("读取zip条目 %s 失败: %w", f.Name, err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return 0, fmt.Errorf("写入文件 %s 失败: %w", fullPath, err)
		}
		rc.Close()
		outFile.Close()
		count++
	}
	return count, nil
}
