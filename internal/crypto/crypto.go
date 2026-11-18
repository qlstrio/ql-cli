package crypto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

type File struct {
	AbsPath string
	RelPath string
}

type Options struct {
	Source     string
	Out        string
	Recursive  bool
	Filter     []string
	KeepDir    bool
	Suffix     string
	Recipient  string
	Identity   string
	Passphrase string
}

func Keygen(outPath string) (string, error) {
	identity, err := generateHybridIdentity()
	if err != nil {
		return "", fmt.Errorf("生成密钥对失败: %w", err)
	}

	pub := identity.Recipient().String()
	priv := identity.String()

	if outPath == "" {
		home, _ := os.UserHomeDir()
		outPath = filepath.Join(home, ".age", "key.txt")
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
		return "", fmt.Errorf("创建密钥目录失败: %w", err)
	}
	if err := os.WriteFile(outPath, []byte(priv+"\n"), 0600); err != nil {
		return "", fmt.Errorf("写入私钥失败: %w", err)
	}

	return pub, nil
}

func Encrypt(opts Options) error {
	files, isDir, err := CollectFiles(opts.Source, opts.Filter, opts.Recursive)
	if err != nil {
		return err
	}

	if !isDir && opts.Out == "" && opts.Suffix == "" {
		return fmt.Errorf("单文件加密必须指定 -o 输出路径 或 --suffix 后缀, 否则会覆盖源文件")
	}

	return batchProcess(files, isDir, opts.Out, opts.KeepDir, opts.Suffix, true,
		func(in, out string) error {
			return encryptFile(in, out, opts.Recipient, opts.Passphrase)
		})
}

func Decrypt(opts Options) error {
	files, isDir, err := CollectFiles(opts.Source, opts.Filter, opts.Recursive)
	if err != nil {
		return err
	}

	return batchProcess(files, isDir, opts.Out, opts.KeepDir, opts.Suffix, false,
		func(in, out string) error {
			return decryptFile(in, out, opts.Identity, opts.Passphrase)
		})
}

func CollectFiles(source string, filters []string, recursive bool) ([]File, bool, error) {
	absSource, err := filepath.Abs(source)
	if err != nil {
		return nil, false, fmt.Errorf("解析路径失败: %w", err)
	}

	info, err := os.Stat(absSource)
	if err != nil {
		return nil, false, fmt.Errorf("路径不存在: %w", err)
	}

	if !info.IsDir() {
		return []File{{AbsPath: absSource, RelPath: filepath.Base(absSource)}}, false, nil
	}

	var exts map[string]bool
	if len(filters) > 0 {
		exts = make(map[string]bool)
		for _, f := range filters {
			f = strings.ToLower(strings.TrimPrefix(f, "."))
			exts[f] = true
		}
	}

	var files []File
	walk := func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		if exts != nil && !exts[strings.ToLower(strings.TrimPrefix(filepath.Ext(p), "."))] {
			return nil
		}
		rel, _ := filepath.Rel(absSource, p)
		files = append(files, File{AbsPath: p, RelPath: rel})
		return nil
	}

	if recursive {
		err = filepath.Walk(absSource, walk)
	} else {
		entries, readErr := os.ReadDir(absSource)
		if readErr != nil {
			return nil, true, readErr
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			p := filepath.Join(absSource, e.Name())
			if exts != nil && !exts[strings.ToLower(strings.TrimPrefix(filepath.Ext(p), "."))] {
				continue
			}
			rel, _ := filepath.Rel(absSource, p)
			files = append(files, File{AbsPath: p, RelPath: rel})
		}
	}
	if err != nil {
		return nil, true, fmt.Errorf("扫描目录失败: %w", err)
	}
	return files, true, nil
}

func resolveBatchOutPath(f File, isDir bool, outDir string, keepDir bool, suffix string, isEncrypt bool) (string, error) {
	var baseName, relName string
	if isEncrypt {
		baseName = filepath.Base(f.AbsPath) + suffix
		relName = f.RelPath + suffix
	} else {
		baseName = strings.TrimSuffix(filepath.Base(f.AbsPath), suffix)
		relName = strings.TrimSuffix(f.RelPath, suffix)
	}

	if outDir == "" {
		if isEncrypt {
			return f.AbsPath + suffix, nil
		}
		return strings.TrimSuffix(f.AbsPath, suffix), nil
	}

	outInfo, outErr := os.Stat(outDir)

	if outErr == nil && outInfo.IsDir() {
		if isDir && keepDir {
			return filepath.Join(outDir, relName), nil
		}
		return filepath.Join(outDir, baseName), nil
	}

	if outErr == nil {
		return "", fmt.Errorf("%s 已存在且不是目录, 无法作为批量输出路径", outDir)
	}

	if !isDir {
		return outDir, nil
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}
	if keepDir {
		return filepath.Join(outDir, relName), nil
	}
	return filepath.Join(outDir, baseName), nil
}

func batchProcess(files []File, isDir bool, outDir string, keepDir bool, suffix string, isEncrypt bool, fn func(in, out string) error) error {
	for _, f := range files {
		outPath, err := resolveBatchOutPath(f, isDir, outDir, keepDir, suffix, isEncrypt)
		if err != nil {
			fmt.Printf("  ✗ %v\n", color.RedString(err.Error()))
			continue
		}

		fmt.Printf("▸ %s\n    → %s\n", color.CyanString(f.AbsPath), outPath)
		if err := fn(f.AbsPath, outPath); err != nil {
			fmt.Printf("  ✗ %v\n", color.RedString(err.Error()))
			continue
		}
		fmt.Printf("  ✓ 完成\n")
	}
	return nil
}
