package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	t.Run("基本复制", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "sub", "dst.txt")
		content := []byte("hello world")
		os.WriteFile(src, content, 0644)

		if err := CopyFile(src, dst); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(dst)
		if string(got) != string(content) {
			t.Errorf("内容不匹配: got %q, want %q", got, content)
		}
	})

	t.Run("源不存在", func(t *testing.T) {
		err := CopyFile("/nonexistent/file.txt", filepath.Join(t.TempDir(), "out.txt"))
		if err == nil {
			t.Fatal("应该报错")
		}
	})

	t.Run("自动创建目标目录", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		content := []byte("data")
		os.WriteFile(src, content, 0644)

		dst := filepath.Join(dir, "deep", "nested", "subdir", "dst.txt")
		if err := CopyFile(src, dst); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(dst)
		if string(got) != "data" {
			t.Errorf("内容不匹配: %q", got)
		}
	})

	t.Run("目标已存在会覆盖", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		os.WriteFile(src, []byte("new"), 0644)
		os.WriteFile(dst, []byte("old"), 0644)

		CopyFile(src, dst)
		got, _ := os.ReadFile(dst)
		if string(got) != "new" {
			t.Errorf("应该覆盖: got %q", got)
		}
	})

	t.Run("空文件也能复制", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "empty.txt")
		dst := filepath.Join(dir, "out", "empty_copy.txt")
		os.WriteFile(src, []byte{}, 0644)

		if err := CopyFile(src, dst); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(dst)
		if len(got) != 0 {
			t.Errorf("应该是空文件, 实际 %d bytes", len(got))
		}
	})
}

func TestMoveFile(t *testing.T) {
	t.Run("基本移动同目录", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		content := []byte("move me")
		os.WriteFile(src, content, 0644)

		if err := MoveFile(src, dst); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(dst)
		if string(got) != "move me" {
			t.Errorf("内容不匹配: %q", got)
		}
		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Error("源文件应该被删除")
		}
	})

	t.Run("移动到不同子目录自动创建", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "from", "a.txt")
		dst := filepath.Join(dir, "to", "deep", "a.txt")
		os.Mkdir(filepath.Join(dir, "from"), 0755)
		content := []byte("v")
		os.WriteFile(src, content, 0644)

		if err := MoveFile(src, dst); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(dst)
		if string(got) != "v" {
			t.Errorf("内容不匹配: %q", got)
		}
		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Error("源文件应该被删除")
		}
	})

	t.Run("源不存在报错", func(t *testing.T) {
		err := MoveFile("/nonexistent/file.txt", filepath.Join(t.TempDir(), "out.txt"))
		if err == nil {
			t.Fatal("应该报错")
		}
	})
}

func TestCheckDirExist(t *testing.T) {
	t.Run("自动创建深目录", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		os.WriteFile(src, []byte("ok"), 0644)
		deepDst := filepath.Join(dir, "a", "b", "c", "file.txt")

		err := checkDirExist(src, deepDst)
		if err != nil {
			t.Fatal(err)
		}

		info, err := os.Stat(filepath.Dir(deepDst))
		if err != nil {
			t.Fatalf("目录没创建: %v", err)
		}
		if !info.IsDir() {
			t.Error("应该是目录")
		}
	})

	t.Run("源不存在报错", func(t *testing.T) {
		err := checkDirExist("/nonexistent", filepath.Join(t.TempDir(), "out.txt"))
		if err == nil {
			t.Fatal("应该报错")
		}
	})
}
