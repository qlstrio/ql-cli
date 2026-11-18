package splitZip

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestGroupBySize(t *testing.T) {
	t.Run("空输入", func(t *testing.T) {
		groups, err := groupBySize(nil, 100)
		if err != nil {
			t.Fatal(err)
		}
		if len(groups) != 0 {
			t.Errorf("空输入应该返回空分组, 实际 %d", len(groups))
		}
	})

	t.Run("单个文件小于max", func(t *testing.T) {
		files := []FileSize{{Size: 10, Path: "a.txt"}}
		groups, _ := groupBySize(files, 100)
		if len(groups) != 1 || len(groups[0]) != 1 {
			t.Errorf("want [[a.txt]], 实际 %v", groups)
		}
	})

	t.Run("单个文件大于max单独成组", func(t *testing.T) {
		files := []FileSize{{Size: 200, Path: "big.bin"}}
		groups, _ := groupBySize(files, 100)
		if len(groups) != 1 || groups[0][0] != "big.bin" {
			t.Errorf("应该单独成组, 实际 %v", groups)
		}
	})

	t.Run("多个小文件凑组且不超max", func(t *testing.T) {
		files := []FileSize{
			{Size: 50, Path: "a"},
			{Size: 40, Path: "b"},
			{Size: 30, Path: "c"},
			{Size: 20, Path: "d"},
		}
		groups, _ := groupBySize(files, 100)

		if len(groups) != 2 {
			t.Fatalf("应该分成2组, 实际 %v", groups)
		}

		allPaths := []string{}
		for _, g := range groups {
			sum := int64(0)
			allPaths = append(allPaths, g...)
			for _, p := range g {
				for _, f := range files {
					if f.Path == p {
						sum += f.Size
					}
				}
			}
			if sum > 100 {
				t.Errorf("分组大小超过 max: %d (paths=%v)", sum, g)
			}
		}

		sort.Strings(allPaths)
		want := []string{"a", "b", "c", "d"}
		if !reflect.DeepEqual(allPaths, want) {
			t.Errorf("应该覆盖所有文件, got %v", allPaths)
		}
	})

	t.Run("大文件优先打包", func(t *testing.T) {
		files := []FileSize{
			{Size: 5, Path: "tiny"},
			{Size: 80, Path: "big"},
			{Size: 15, Path: "small"},
		}
		groups, _ := groupBySize(files, 100)
		if len(groups) != 1 {
			t.Fatalf("应该1组(80+15+5<100), 实际 %v", groups)
		}
	})

	t.Run("恰好等max和稍超的组合", func(t *testing.T) {
		files := []FileSize{
			{Size: 100, Path: "exact"},
			{Size: 1, Path: "tiny"},
		}
		groups, _ := groupBySize(files, 100)
		if len(groups) != 2 {
			t.Fatalf("exact=100单独一组, tiny=1单独一组, 实际 %v", groups)
		}
	})

	t.Run("先大后小贪心凑组", func(t *testing.T) {
		files := []FileSize{
			{Size: 80, Path: "big"},
			{Size: 25, Path: "med"},
			{Size: 5, Path: "tiny"},
		}
		groups, _ := groupBySize(files, 100)
		if len(groups) != 2 {
			t.Fatalf("80+5=85一组, 25单独一组, 实际 %v", groups)
		}
	})
}

func TestDoZipDoExtract_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	outDir := filepath.Join(dir, "out")
	unzipDir := filepath.Join(dir, "unzip")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(outDir, 0755)
	os.MkdirAll(unzipDir, 0755)

	files := map[string]string{
		"a.txt":      "hello world",
		"sub/b.json": `{"key":"value"}`,
		"sub/c.bin":  string(make([]byte, 200)),
		"top.md":     "# markdown",
	}

	for name, content := range files {
		p := filepath.Join(srcDir, name)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte(content), 0644)
	}

	DoZip(srcDir, outDir, 150, "part_")

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Fatalf("应该分割成至少2个zip, 实际 %d", len(entries))
	}

	zipPaths := []string{}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".zip" {
			zipPaths = append(zipPaths, filepath.Join(outDir, e.Name()))
		}
	}

	if err := DoExtract(zipPaths, unzipDir); err != nil {
		t.Fatalf("DoExtract error: %v", err)
	}

	for name, content := range files {
		got, err := os.ReadFile(filepath.Join(unzipDir, name))
		if err != nil {
			t.Errorf("解压后缺文件: %s (%v)", name, err)
			continue
		}
		if string(got) != content {
			t.Errorf("文件 %s 内容不匹配", name)
		}
	}
}

func TestExtractZip_MissingZip(t *testing.T) {
	_, err := extractZip("/nonexistent.zip", t.TempDir())
	if err == nil {
		t.Fatal("应该报错")
	}
}

func TestExtractZip_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	evilZip := filepath.Join(dir, "evil.zip")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("../../evil.txt")
	w.Write([]byte("pwned"))
	zw.Close()

	os.WriteFile(evilZip, buf.Bytes(), 0644)

	outDir := filepath.Join(dir, "safe_out")
	_, err := extractZip(evilZip, outDir)
	if err == nil {
		t.Fatal("路径穿越应该被拦截")
	}
	if !contains(err.Error(), "非法路径") {
		t.Errorf("错误信息应该包含非法路径, 实际: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
