package fileops

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseExtensions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]bool
	}{
		{"nil输入", nil, nil},
		{"空切片", []string{}, nil},
		{"单个无点", []string{"txt"}, map[string]bool{"txt": true}},
		{"单个带点", []string{".JSON"}, map[string]bool{"json": true}},
		{"逗号分隔", []string{"txt,json"}, map[string]bool{"txt": true, "json": true}},
		{"多flag+逗号混合", []string{"txt", "json,bin"}, map[string]bool{"txt": true, "json": true, "bin": true}},
		{"空格过滤", []string{" txt ", " .md "}, map[string]bool{"txt": true, "md": true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExtensions(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("parseExtensions(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCommonOptionsValidate(t *testing.T) {
	dir := t.TempDir()

	t.Run("空Dir", func(t *testing.T) {
		err := CommonOptions{}.Validate()
		if err == nil {
			t.Error("应该报错")
		}
	})

	t.Run("源=目标", func(t *testing.T) {
		err := CommonOptions{Dir: dir, OutDir: dir}.Validate()
		if err == nil {
			t.Error("应该报错")
		}
	})

	t.Run("合法无OutDir", func(t *testing.T) {
		err := CommonOptions{Dir: dir}.Validate()
		if err != nil {
			t.Errorf("不该报错: %v", err)
		}
	})

	t.Run("合法有OutDir", func(t *testing.T) {
		err := CommonOptions{Dir: dir, OutDir: filepath.Join(dir, "out")}.Validate()
		if err != nil {
			t.Errorf("不该报错: %v", err)
		}
	})
}

func TestCollectFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.json"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "c.bin"), []byte("c"), 0644)
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "d.md"), []byte("d"), 0644)

	t.Run("非递归", func(t *testing.T) {
		files, err := CollectFiles(dir, nil, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 3 {
			t.Fatalf("want 3, got %d", len(files))
		}
	})

	t.Run("递归", func(t *testing.T) {
		files, err := CollectFiles(dir, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 4 {
			t.Fatalf("want 4, got %d", len(files))
		}
	})

	t.Run("递归+筛选", func(t *testing.T) {
		files, err := CollectFiles(dir, []string{"txt", "json"}, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 2 {
			t.Fatalf("want 2, got %d", len(files))
		}
	})

	t.Run("非递归+筛选", func(t *testing.T) {
		files, err := CollectFiles(dir, []string{"txt"}, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 1 {
			t.Fatalf("want 1, got %d", len(files))
		}
	})

	t.Run("不存在目录", func(t *testing.T) {
		_, err := CollectFiles("/nonexistent/path", nil, false)
		if err == nil {
			t.Error("应该报错")
		}
	})

	t.Run("不是目录", func(t *testing.T) {
		_, err := CollectFiles(filepath.Join(dir, "a.txt"), nil, false)
		if err == nil {
			t.Error("应该报错")
		}
	})

	t.Run("空目录", func(t *testing.T) {
		empty := t.TempDir()
		files, err := CollectFiles(empty, nil, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 0 {
			t.Errorf("空目录应该0个文件, 实际 %d", len(files))
		}
	})
}

func TestResolveOutPath(t *testing.T) {
	dir := t.TempDir()
	f := File{AbsPath: filepath.Join(dir, "sub", "a.txt"), RelPath: filepath.Join("sub", "a.txt")}

	t.Run("keepDir保留结构", func(t *testing.T) {
		out, err := ResolveOutPath(dir, filepath.Join(dir, "out"), true, f)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "out", "sub", "a.txt")
		if out != want {
			t.Errorf("got %q, want %q", out, want)
		}
	})

	t.Run("不keepDir只用basename", func(t *testing.T) {
		out, err := ResolveOutPath(dir, filepath.Join(dir, "out2"), false, f)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "out2", "a.txt")
		if out != want {
			t.Errorf("got %q, want %q", out, want)
		}
	})

	t.Run("源=目标报错", func(t *testing.T) {
		_, err := ResolveOutPath(dir, dir, false, f)
		if err == nil {
			t.Error("应该报错")
		}
	})
}

func TestRenameValidateRegex(t *testing.T) {
	t.Run("合法正则", func(t *testing.T) {
		err := RenameOptions{UseRegex: true, From: `IMG_(\d+)`}.ValidateRegex()
		if err != nil {
			t.Errorf("不该报错: %v", err)
		}
	})

	t.Run("非法正则", func(t *testing.T) {
		err := RenameOptions{UseRegex: true, From: "["}.ValidateRegex()
		if err == nil {
			t.Error("应该报错")
		}
	})

	t.Run("不使用正则跳过", func(t *testing.T) {
		err := RenameOptions{UseRegex: false, From: "["}.ValidateRegex()
		if err != nil {
			t.Errorf("不该报错: %v", err)
		}
	})
}

func TestRenameComputeName(t *testing.T) {
	tests := []struct {
		name  string
		opts  RenameOptions
		input string
		want  string
	}{
		{"无变换", RenameOptions{}, "file.txt", "file.txt"},
		{"查找替换", RenameOptions{From: "IMG", To: "Photo", UseRegex: false}, "IMG_001.txt", "Photo_001.txt"},
		{"正则替换", RenameOptions{From: `IMG_(\d+)`, To: "Photo_$1", UseRegex: true}, "IMG_001.txt", "Photo_001.txt"},
		{"添加前缀", RenameOptions{Prefix: "2024_"}, "report.pdf", "2024_report.pdf"},
		{"添加后缀", RenameOptions{Suffix: "_final"}, "data.csv", "data_final.csv"},
		{"后缀在最后一个扩展名前", RenameOptions{Suffix: "_v2"}, "data.tar.gz", "data.tar_v2.gz"},
		{"转小写", RenameOptions{ToLower: true}, "FILE.TXT", "file.TXT"},
		{"转大写", RenameOptions{ToUpper: true}, "file.txt", "FILE.txt"},
		{"去扩展名", RenameOptions{RemoveExt: true}, "archive.zip", "archive"},
		{"清理特殊字符", RenameOptions{Clean: true}, "bad@file#name.txt", "badfilename.txt"},
		{"清理保留中文", RenameOptions{Clean: true}, "报告-2024.txt", "报告-2024.txt"},
		{"组合: 前缀+小写", RenameOptions{Prefix: "PRE_", ToLower: true}, "FILE.TXT", "PRE_file.TXT"},
		{"组合: 前缀+替换+后缀", RenameOptions{Prefix: "A_", From: "IMG", To: "PH", UseRegex: false, Suffix: "_Z"}, "IMG_001.txt", "A_PH_001_Z.txt"},
		{"正则替换失败时原值返回", RenameOptions{UseRegex: true, From: "["}, "file.txt", "file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.computeName(tt.input)
			if got != tt.want {
				t.Errorf("computeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRenameActual(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "IMG_001.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "IMG_002.txt"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "keep.json"), []byte("c"), 0644)

	t.Run("基本查找替换", func(t *testing.T) {
		err := RenameOptions{
			Dir:      dir,
			From:     "IMG_",
			To:       "Photo_",
			UseRegex: false,
			Filter:   []string{"txt"},
		}.Rename()
		if err != nil {
			t.Fatal(err)
		}

		files, _ := os.ReadDir(dir)
		names := make(map[string]bool)
		for _, f := range files {
			names[f.Name()] = true
		}
		if !names["Photo_001.txt"] || !names["Photo_002.txt"] {
			t.Errorf("期望 Photo_001.txt 和 Photo_002.txt, 实际: %v", names)
		}
		if !names["keep.json"] {
			t.Error("keep.json 不应该被改名")
		}
	})
}

func TestRenameDryRun(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "IMG_001.txt"), []byte("a"), 0644)

	err := RenameOptions{
		Dir:      dir,
		From:     "IMG",
		To:       "Photo",
		DryRun:   true,
		UseRegex: false,
	}.Rename()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "IMG_001.txt")); err != nil {
		t.Error("dry-run 模式文件不应被实际修改")
	}
}

func TestRenameRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(dir, "top_keep.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(sub, "sub_keep.txt"), []byte("b"), 0644)

	err := RenameOptions{
		Dir:       dir,
		Recursive: true,
		From:      "keep",
		To:        "changed",
		UseRegex:  false,
	}.Rename()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "top_changed.txt")); err != nil {
		t.Error("顶层文件应该被重命名")
	}
	if _, err := os.Stat(filepath.Join(sub, "sub_changed.txt")); err != nil {
		t.Error("子目录文件应该被重命名")
	}
}

func TestRenameNoMatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0644)

	err := RenameOptions{
		Dir:  dir,
		From: "不存在的字符串",
		To:   "yyy",
	}.Rename()
	if err != nil {
		t.Fatal(err)
	}

	files, _ := os.ReadDir(dir)
	if len(files) != 1 || files[0].Name() != "keep.txt" {
		t.Errorf("无匹配时不应改变任何文件, 实际: %v", files)
	}
}

func TestBatch(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.json"), []byte("b"), 0644)

	t.Run("全部成功", func(t *testing.T) {
		called := make([]string, 0)
		err := Batch(CommonOptions{
			Dir:       dir,
			OutDir:    outDir,
			Recursive: false,
		}, func(src, dst string) error {
			called = append(called, src)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(called) != 2 {
			t.Errorf("应该调用2次, 实际 %d", len(called))
		}
	})

	t.Run("单个失败不中断", func(t *testing.T) {
		called := make([]string, 0)
		err := Batch(CommonOptions{
			Dir:       dir,
			OutDir:    filepath.Join(dir, "out2"),
			Recursive: false,
		}, func(src, dst string) error {
			called = append(called, src)
			if strings.HasSuffix(src, ".txt") {
				return os.ErrPermission
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(called) != 2 {
			t.Errorf("应该遍历所有文件, 实际调用 %d 次", len(called))
		}
	})
}
