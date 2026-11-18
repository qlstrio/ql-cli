package splitZip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type FileSize struct {
	Size int64
	Path string
}

type ByLargest []FileSize

func (f ByLargest) Len() int           { return len(f) }
func (f ByLargest) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByLargest) Less(i, j int) bool { return f[i].Size > f[j].Size }

func DoZip(sourceFilePath string, outputPath string, maxSize int64, prefix string) {
	var fileSize []FileSize
	filepath.Walk(sourceFilePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		f := FileSize{
			Size: info.Size(),
			Path: path,
		}
		fileSize = append(fileSize, f)
		return nil
	})
	totoalFileSize := len(fileSize)
	ret, _ := groupBySize(fileSize, maxSize)
	currFileSize := 0
	for index, stringInfo := range ret {
		currFileSize += len(stringInfo)
		if err := writeZip(stringInfo, fmt.Sprintf("%v%v.zip", prefix, index), outputPath, sourceFilePath); err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println(totoalFileSize, currFileSize)
}

func groupBySize(files []FileSize, maxSize int64) ([][]string, error) {
	var ret [][]string

	sort.Sort(ByLargest(files))
	for i, e := range files {
		if e.Size == -1 {
			continue
		}
		if e.Size > maxSize {
			// 单个文件大小超过最大文件大小，单独压缩
			ret = append(ret, []string{e.Path})
			continue
		}
		files[i].Size = -1

		group := []string{e.Path}
		d := maxSize - e.Size

		j := i + 1
		s := e.Size
		for ; j < len(files); j++ {
			if files[j].Size == -1 {
				continue
			}
			if files[j].Size <= d {
				s += files[j].Size
				d = maxSize - s
				group = append(group, files[j].Path)
				files[j].Size = -1
			}
			if s > maxSize {
				break
			}
		}
		ret = append(ret, group)
	}
	return ret, nil
}

func writeZip(filePaths []string, zipName string, outputPath string, sourceFilePath string) error {
	zipPath := filepath.Join(outputPath, zipName)
	f, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file %s: %w", zipPath, err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, filePath := range filePaths {
		relPath, err := filepath.Rel(sourceFilePath, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
		}
		relPath = filepath.ToSlash(relPath)

		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", filePath, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", filePath, err)
		}
		header.Name = relPath
		header.Method = zip.Deflate

		w, err := zw.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip entry for %s: %w", filePath, err)
		}

		srcFile, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filePath, err)
		}

		if _, err := io.Copy(w, srcFile); err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to write file %s to zip: %w", filePath, err)
		}
		srcFile.Close()
	}
	return nil
}
