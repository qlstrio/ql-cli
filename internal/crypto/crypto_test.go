package crypto

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeygen(t *testing.T) {
	t.Run("默认路径", func(t *testing.T) {
		pub, err := Keygen("")
		if err != nil {
			t.Fatalf("Keygen() error = %v", err)
		}
		if !strings.HasPrefix(pub, "age1") {
			t.Errorf("公钥格式不对: %s", pub)
		}
	})

	t.Run("指定输出路径+自动创建子目录", func(t *testing.T) {
		dir := t.TempDir()
		out := filepath.Join(dir, "sub", "nested", "key.txt")
		pub, err := Keygen(out)
		if err != nil {
			t.Fatalf("Keygen() error = %v", err)
		}
		if _, err := os.Stat(out); err != nil {
			t.Fatalf("私钥文件未生成: %v", err)
		}
		data, _ := os.ReadFile(out)
		if !strings.Contains(string(data), "AGE-SECRET-KEY") {
			t.Errorf("私钥内容不对")
		}
		if !strings.HasPrefix(pub, "age1") {
			t.Errorf("公钥格式不对: %s", pub)
		}
	})
}

func TestEncryptDecrypt_Passphrase(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "secret.txt")
	content := []byte("这是一段机密数据！")
	os.WriteFile(src, content, 0644)

	enc := filepath.Join(dir, "secret.txt.age")
	dec := filepath.Join(dir, "secret.dec")

	if err := encryptFile(src, enc, "", "密码"); err != nil {
		t.Fatalf("encryptFile() error = %v", err)
	}
	encData, _ := os.ReadFile(enc)
	if bytes.Equal(encData, content) {
		t.Fatal("加密文件内容与原文相同")
	}

	if err := decryptFile(enc, dec, "", "密码"); err != nil {
		t.Fatalf("decryptFile() error = %v", err)
	}
	decData, _ := os.ReadFile(dec)
	if !bytes.Equal(decData, content) {
		t.Fatalf("解密内容不匹配: got %q, want %q", decData, content)
	}

	if err := decryptFile(enc, dec+".2", "", "错误密码"); err == nil {
		t.Fatal("错误密码应该解密失败")
	}
}

func TestEncryptDecrypt_KeyPair(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.txt")
	pub, err := Keygen(keyFile)
	if err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(dir, "secret.txt")
	content := []byte("非对称加密数据")
	os.WriteFile(src, content, 0644)

	enc := filepath.Join(dir, "secret.age")
	dec := filepath.Join(dir, "secret.dec")

	if err := encryptFile(src, enc, pub, ""); err != nil {
		t.Fatalf("encryptFile() error = %v", err)
	}
	if err := decryptFile(enc, dec, keyFile, ""); err != nil {
		t.Fatalf("decryptFile() error = %v", err)
	}
	decData, _ := os.ReadFile(dec)
	if !bytes.Equal(decData, content) {
		t.Fatalf("解密内容不匹配: got %q", decData)
	}

	otherKey := filepath.Join(dir, "other.txt")
	Keygen(otherKey)
	if err := decryptFile(enc, dec+".2", otherKey, ""); err == nil {
		t.Fatal("用错误私钥应该解密失败")
	}
}

func TestEncrypt_RecipientAndPassphraseConflict(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.txt")
	pub, _ := Keygen(keyFile)

	src := filepath.Join(dir, "data.txt")
	os.WriteFile(src, []byte("x"), 0644)

	err := encryptFile(src, filepath.Join(dir, "x.age"), pub, "额外密码")
	if err == nil {
		t.Fatal("混合 post-quantum HybridIdentity 和 classic Scrypt 应该报错")
	}
	if !strings.Contains(err.Error(), "incompatible") {
		t.Errorf("错误信息应该提到 incompatible: %v", err)
	}
}

func TestEncrypt_InvalidRecipient(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	os.WriteFile(src, []byte("x"), 0644)

	err := encryptFile(src, filepath.Join(dir, "a.age"), "age1invalid!!!", "")
	if err == nil {
		t.Fatal("非法 recipient 应该报错")
	}
	if !strings.Contains(err.Error(), "解析公钥失败") {
		t.Errorf("错误信息不对: %v", err)
	}
}

func TestDecrypt_SourceNotExist(t *testing.T) {
	err := decryptFile("/nonexistent/file.age", "/tmp/out", "", "pw")
	if err == nil {
		t.Fatal("源不存在应该报错")
	}
}

func TestDecrypt_IdentityNotExist(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.age")
	os.WriteFile(src, []byte("fake age data"), 0644)

	err := decryptFile(src, filepath.Join(dir, "out"), "/nonexistent/key.txt", "")
	if err == nil {
		t.Fatal("私钥文件不存在应该报错")
	}
	if !strings.Contains(err.Error(), "读取私钥文件失败") {
		t.Errorf("错误信息不对: %v", err)
	}
}

func TestEncrypt_BatchDir(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "enc_out")

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(dir, "b.json"), []byte("bbb"), 0644)
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "c.md"), []byte("ccc"), 0644)

	err := Encrypt(Options{
		Source:     dir,
		Out:        outDir,
		Recursive:  true,
		KeepDir:    true,
		Suffix:     ".age",
		Passphrase: "batchpw",
	})
	if err != nil {
		t.Fatalf("Encrypt batch error = %v", err)
	}

	t.Run("文件数正确", func(t *testing.T) {
		var encrypted []string
		filepath.Walk(outDir, func(p string, fi os.FileInfo, err error) error {
			if !fi.IsDir() {
				encrypted = append(encrypted, filepath.Base(p))
			}
			return nil
		})
		if len(encrypted) != 3 {
			t.Fatalf("应该有3个加密文件, 实际 %v", encrypted)
		}
	})

	t.Run("解密回原内容", func(t *testing.T) {
		decDir := filepath.Join(dir, "dec_out")
		err := Decrypt(Options{
			Source:     outDir,
			Out:        decDir,
			Recursive:  true,
			KeepDir:    true,
			Suffix:     ".age",
			Passphrase: "batchpw",
		})
		if err != nil {
			t.Fatal(err)
		}

		a, _ := os.ReadFile(filepath.Join(decDir, "a.txt"))
		c, _ := os.ReadFile(filepath.Join(decDir, "sub", "c.md"))
		if string(a) != "aaa" {
			t.Errorf("a.txt 内容不对: %q", a)
		}
		if string(c) != "ccc" {
			t.Errorf("sub/c.md 内容不对: %q", c)
		}
	})
}

func TestEncrypt_BatchDir_Filter(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "enc_out")

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(dir, "b.json"), []byte("bbb"), 0644)

	err := Encrypt(Options{
		Source:     dir,
		Out:        outDir,
		Recursive:  false,
		Filter:     []string{"txt"},
		Suffix:     ".age",
		Passphrase: "f",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "a.txt.age")); err != nil {
		t.Error("a.txt.age 应该存在")
	}
	if _, err := os.Stat(filepath.Join(outDir, "b.json.age")); !os.IsNotExist(err) {
		t.Error("b.json.age 不应该被加密")
	}
}

func TestEncrypt_SingleFile_NoOut_NoSuffix(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "test.txt")
	os.WriteFile(src, []byte("data"), 0644)

	err := Encrypt(Options{Source: src, Passphrase: "pw"})
	if err == nil {
		t.Fatal("单文件加密无 -o 无 suffix 应该报错(防止覆盖)")
	}
	if !strings.Contains(err.Error(), "覆盖") {
		t.Errorf("错误信息应该提到覆盖, 实际: %v", err)
	}
}

func TestCollectFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.json"), []byte("b"), 0644)
	os.WriteFile(filepath.Join(dir, "c.bin"), []byte("c"), 0644)
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "d.md"), []byte("d"), 0644)

	t.Run("单文件", func(t *testing.T) {
		files, isDir, err := CollectFiles(filepath.Join(dir, "a.txt"), nil, false)
		if err != nil {
			t.Fatal(err)
		}
		if isDir {
			t.Error("应该是单文件")
		}
		if len(files) != 1 {
			t.Fatalf("文件数不对: %d", len(files))
		}
	})

	t.Run("目录非递归", func(t *testing.T) {
		files, isDir, err := CollectFiles(dir, nil, false)
		if err != nil {
			t.Fatal(err)
		}
		if !isDir {
			t.Error("应该是目录")
		}
		if len(files) != 3 {
			t.Fatalf("应该只有3个根文件(非递归), 实际 %d", len(files))
		}
	})

	t.Run("目录递归", func(t *testing.T) {
		files, _, err := CollectFiles(dir, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 4 {
			t.Fatalf("递归应该4个文件, 实际 %d", len(files))
		}
		hasSub := false
		for _, f := range files {
			if strings.Contains(f.AbsPath, "sub") {
				hasSub = true
				break
			}
		}
		if !hasSub {
			t.Error("递归未包含子目录文件")
		}
	})

	t.Run("按后缀筛选", func(t *testing.T) {
		files, _, err := CollectFiles(dir, []string{"txt", "json"}, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 2 {
			t.Fatalf("筛选后应该2个文件, 实际 %d", len(files))
		}
	})

	t.Run("不存在路径", func(t *testing.T) {
		_, _, err := CollectFiles("/nonexistent/path", nil, false)
		if err == nil {
			t.Error("应该返回错误")
		}
	})
}

func TestResolveBatchOutPath(t *testing.T) {
	tmpRoot := t.TempDir()
	f := File{AbsPath: filepath.Join(tmpRoot, "sub", "a.txt"), RelPath: filepath.Join("sub", "a.txt")}

	t.Run("加密到源同目录无outDir", func(t *testing.T) {
		out, err := resolveBatchOutPath(f, true, "", false, ".age", true)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(tmpRoot, "sub", "a.txt.age")
		if out != want {
			t.Errorf("got %q, want %q", out, want)
		}
	})

	t.Run("解密去掉后缀到outDir保留结构", func(t *testing.T) {
		outDir := filepath.Join(tmpRoot, "out")
		f2 := File{AbsPath: filepath.Join(tmpRoot, "sub", "a.txt.age"), RelPath: filepath.Join("sub", "a.txt.age")}
		out, err := resolveBatchOutPath(f2, true, outDir, true, ".age", false)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(outDir, "sub", "a.txt")
		if out != want {
			t.Errorf("got %q, want %q", out, want)
		}
	})

	t.Run("加密到outDir不保留结构", func(t *testing.T) {
		outDir := filepath.Join(tmpRoot, "out2")
		out, err := resolveBatchOutPath(f, true, outDir, false, ".age", true)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(outDir, "a.txt.age")
		if out != want {
			t.Errorf("got %q, want %q", out, want)
		}
	})

	t.Run("outDir已存在且是目录→保留结构", func(t *testing.T) {
		outDir := filepath.Join(tmpRoot, "exist_out")
		os.MkdirAll(outDir, 0755)
		out, err := resolveBatchOutPath(f, true, outDir, true, ".age", true)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(outDir, "sub", "a.txt.age")
		if out != want {
			t.Errorf("got %q, want %q", out, want)
		}
	})

	t.Run("outDir已存在且是文件→报错", func(t *testing.T) {
		outFile := filepath.Join(tmpRoot, "somefile")
		os.WriteFile(outFile, []byte("x"), 0644)
		_, err := resolveBatchOutPath(f, true, outFile, false, ".age", true)
		if err == nil {
			t.Fatal("应该报错: 已存在且不是目录")
		}
		if !strings.Contains(err.Error(), "不是目录") {
			t.Errorf("错误信息不对: %v", err)
		}
	})

	t.Run("非单文件加密→outDir直接作为输出路径", func(t *testing.T) {
		outDir := filepath.Join(tmpRoot, "single_out")
		out, err := resolveBatchOutPath(f, false, outDir, false, ".age", true)
		if err != nil {
			t.Fatal(err)
		}
		if out != outDir {
			t.Errorf("got %q, want %q", out, outDir)
		}
	})
}
