package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var splitzipCmd = &cobra.Command{
	Use:   "splitzip",
	Short: color.CyanString("将文件分割压缩为多个zip文件"),
	Long:  color.CyanString("将文件分割压缩为多个zip文件，保留文件结构。"),
	RunE:  runSplitZip,
}
var (
	zipfolder    string
	outputfolder string
	zprefix      string
	maxSize      int64
)

func init() {
	splitzipCmd.Flags().StringVarP(&zipfolder, "source", "s", "", "源文件目录")
	splitzipCmd.Flags().StringVarP(&outputfolder, "dir", "d", ".", "输出目录")
	splitzipCmd.Flags().StringVarP(&zprefix, "prefix", "p", "split_", "zip文件前缀")
	splitzipCmd.Flags().Int64VarP(&maxSize, "max", "b", 2, "最大文件大小，单位MB，若单文件大小超过最大文件大小，单独压缩")
	rootCmd.AddCommand(splitzipCmd)
}

func runSplitZip(cmd *cobra.Command, args []string) error {
	var err error
	zipfolder, err = getAbsDirPath(zipfolder)
	if err != nil {
		return err
	}
	outputfolder, err = getAbsDirPath(outputfolder)
	if err != nil {
		return err
	}
	DoZip(zipfolder, outputfolder, maxSize*1024*1024, zprefix)
	return nil
}

type FileSize struct {
	Size int64
	Path string
}

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

type ByLargest []FileSize

func (f ByLargest) Len() int           { return len(f) }
func (f ByLargest) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByLargest) Less(i, j int) bool { return f[i].Size > f[j].Size }

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

var unsplitzipCmd = &cobra.Command{
	Use:   "unsplitzip",
	Short: color.CyanString("将多个分割的zip文件还原为原始目录"),
	Long:  color.CyanString("将 splitzip 生成的多个 zip 文件解压到同一目录，自动合并还原文件结构。"),
	RunE:  runUnsplitZip,
}

var (
	unzipSource string
	unzipOutput string
	unzipPrefix string
)

func init() {
	unsplitzipCmd.Flags().StringVarP(&unzipSource, "source", "s", "", "包含分割zip文件的目录")
	unsplitzipCmd.Flags().StringVarP(&unzipOutput, "dir", "d", ".", "还原输出目录")
	unsplitzipCmd.Flags().StringVarP(&unzipPrefix, "prefix", "p", "split_", "zip文件前缀(与splitzip一致)")
	rootCmd.AddCommand(unsplitzipCmd)
}

func runUnsplitZip(cmd *cobra.Command, args []string) error {
	var err error
	unzipSource, err = getAbsDirPath(unzipSource)
	if err != nil {
		return err
	}
	unzipOutput, err = filepath.Abs(unzipOutput)
	if err != nil {
		return fmt.Errorf("获取输出目录绝对路径失败: %w", err)
	}

	pattern := filepath.Join(unzipSource, unzipPrefix+"*.zip")
	zipFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("查找zip文件失败: %w", err)
	}
	if len(zipFiles) == 0 {
		return fmt.Errorf("在 %s 下未找到匹配 %s*.zip 的文件", unzipSource, unzipPrefix)
	}

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
