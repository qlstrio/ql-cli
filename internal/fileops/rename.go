package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RenameOptions struct {
	Dir       string
	Recursive bool
	Filter    []string
	DryRun    bool

	From      string
	To        string
	UseRegex  bool
	Prefix    string
	Suffix    string
	ToLower   bool
	ToUpper   bool
	Clean     bool
	RemoveExt bool
}

var renameWhitelistPattern = regexp.MustCompile(`[^0-9a-zA-Z_\x{4e00}-\x{9fa5}()（）\-]`)

func (opts RenameOptions) ValidateRegex() error {
	if opts.UseRegex && opts.From != "" {
		if _, err := regexp.Compile(opts.From); err != nil {
			return fmt.Errorf("正则表达式无效: %w", err)
		}
	}
	return nil
}

func (opts RenameOptions) Rename() error {
	if err := opts.ValidateRegex(); err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Println("===== 预览模式(不会实际修改) =====")
	}

	files, err := CollectFiles(opts.Dir, opts.Filter, opts.Recursive)
	if err != nil {
		return err
	}

	count := 0
	for _, f := range files {
		name := filepath.Base(f.AbsPath)
		newName := opts.computeName(name)

		if newName == name {
			continue
		}

		newPath := filepath.Join(filepath.Dir(f.AbsPath), newName)

		if opts.DryRun {
			fmt.Printf("  %s -> %s\n", f.AbsPath, newPath)
		} else {
			if err := os.Rename(f.AbsPath, newPath); err != nil {
				fmt.Fprintf(os.Stderr, "重命名失败 %s: %v\n", f.AbsPath, err)
				continue
			}
			fmt.Printf("已重命名: %s -> %s\n", f.AbsPath, newPath)
		}
		count++
	}

	if count == 0 {
		fmt.Println("没有匹配的文件需要修改")
	} else if opts.DryRun {
		fmt.Printf("共 %d 个文件将被修改\n", count)
	} else {
		fmt.Printf("共修改了 %d 个文件\n", count)
	}
	return nil
}

func (opts RenameOptions) computeName(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	newBase := base

	if opts.Clean {
		newBase = renameWhitelistPattern.ReplaceAllString(newBase, "")
	}

	if opts.RemoveExt {
		ext = ""
	}

	if opts.From != "" {
		if opts.UseRegex {
			re, err := regexp.Compile(opts.From)
			if err == nil {
				newBase = re.ReplaceAllString(newBase, opts.To)
			}
		} else {
			newBase = strings.ReplaceAll(newBase, opts.From, opts.To)
		}
	}

	if opts.ToLower {
		newBase = strings.ToLower(newBase)
	}
	if opts.ToUpper {
		newBase = strings.ToUpper(newBase)
	}

	if opts.Prefix != "" {
		newBase = opts.Prefix + newBase
	}
	if opts.Suffix != "" {
		newBase = newBase + opts.Suffix
	}

	return newBase + ext
}
