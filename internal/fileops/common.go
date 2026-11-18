package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	AbsPath string
	RelPath string
}

type CommonOptions struct {
	Dir       string
	OutDir    string
	Recursive bool
	Filter    []string
	KeepDir   bool
	DryRun    bool
}

func (opts CommonOptions) Validate() error {
	if opts.Dir == "" {
		return fmt.Errorf("请通过 -d 指定源目录")
	}
	if opts.OutDir != "" {
		absDir, err := filepath.Abs(opts.Dir)
		if err != nil {
			return fmt.Errorf("解析源目录失败: %w", err)
		}
		outAbs, err := filepath.Abs(opts.OutDir)
		if err != nil {
			return fmt.Errorf("解析输出目录失败: %w", err)
		}
		if absDir == outAbs {
			return fmt.Errorf("目标目录不能与源目录相同")
		}
	}
	return nil
}

func CollectFiles(dir string, filters []string, recursive bool) ([]File, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("获取目录绝对路径失败: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("目录不存在: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("路径不是目录: %s", absDir)
	}

	exts := parseExtensions(filters)

	var files []File
	walk := func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if p != absDir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if len(exts) > 0 {
			ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(p), "."))
			if !exts[ext] {
				return nil
			}
		}
		rel, _ := filepath.Rel(absDir, p)
		files = append(files, File{AbsPath: p, RelPath: rel})
		return nil
	}

	if err := filepath.Walk(absDir, walk); err != nil {
		return nil, fmt.Errorf("遍历目录失败: %w", err)
	}

	return files, nil
}

func parseExtensions(filters []string) map[string]bool {
	if len(filters) == 0 {
		return nil
	}
	exts := make(map[string]bool)
	for _, f := range filters {
		for _, part := range strings.Split(f, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			part = strings.ToLower(strings.TrimPrefix(part, "."))
			exts[part] = true
		}
	}
	return exts
}

func ResolveOutPath(absDir, outDir string, keepDir bool, f File) (string, error) {
	if outDir == "" {
		outDir = "."
	}
	if !filepath.IsAbs(outDir) {
		twd, _ := os.Getwd()
		outDir = filepath.Join(twd, outDir)
	}

	if outDir == absDir {
		return "", fmt.Errorf("输出目录不能与源目录相同")
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	if keepDir {
		return filepath.Join(outDir, f.RelPath), nil
	}
	return filepath.Join(outDir, filepath.Base(f.AbsPath)), nil
}

func Batch(opts CommonOptions, fn func(src, dst string) error) error {
	files, err := CollectFiles(opts.Dir, opts.Filter, opts.Recursive)
	if err != nil {
		return err
	}

	absDir, _ := filepath.Abs(opts.Dir)

	count := 0
	for _, f := range files {
		dst, err := ResolveOutPath(absDir, opts.OutDir, opts.KeepDir, f)
		if err != nil {
			fmt.Printf("  ✗ %v\n", err)
			continue
		}

		if opts.DryRun {
			fmt.Printf("  %s -> %s\n", f.AbsPath, dst)
		} else {
			if err := fn(f.AbsPath, dst); err != nil {
				fmt.Printf("  ✗ %v\n", err)
				continue
			}
		}
		count++
	}

	if opts.DryRun {
		fmt.Printf("共 %d 个文件将被处理\n", count)
	} else {
		fmt.Printf("共处理了 %d 个文件\n", count)
	}
	return nil
}
