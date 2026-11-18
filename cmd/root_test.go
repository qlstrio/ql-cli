package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("execute() error: %v", err)
	}
	if !strings.Contains(outBuf.String(), "ql-cli") {
		t.Errorf("help output 应该包含 ql-cli, 实际: %q", outBuf.String())
	}
}

func TestCryptoEncrypt_ParamValidation(t *testing.T) {
	cases := []struct {
		name        string
		setupVars   func()
		args        []string
		wantErrMsg  string
		wantNoError bool
	}{
		{
			name:       "缺少-s",
			setupVars:  func() { cryptoSource = ""; cryptoRecipient = ""; cryptoPassphrase = "" },
			args:       []string{},
			wantErrMsg: "-s 指定要加密",
		},
		{
			name:       "缺少 -r 和 -p",
			setupVars:  func() { cryptoSource = "/tmp/a.txt"; cryptoRecipient = ""; cryptoPassphrase = "" },
			args:       []string{},
			wantErrMsg: "-r 指定公钥 或 -p 指定密码",
		},
		{
			name:       "两者都提供",
			setupVars:  func() { cryptoSource = ""; cryptoRecipient = "age1bad"; cryptoPassphrase = "pw" },
			args:       []string{},
			wantErrMsg: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.setupVars()
			err := runCryptoEncrypt(nil, c.args)
			if c.wantErrMsg == "" {
				return
			}
			if err == nil {
				t.Fatalf("期望报错包含 %q, 实际无错误", c.wantErrMsg)
			}
			if !strings.Contains(err.Error(), c.wantErrMsg) {
				t.Errorf("错误信息应该包含 %q, 实际: %v", c.wantErrMsg, err)
			}
		})
	}
}

func TestCryptoDecrypt_ParamValidation(t *testing.T) {
	cases := []struct {
		name       string
		setupVars  func()
		args       []string
		wantErrMsg string
	}{
		{
			name:       "缺少-s",
			setupVars:  func() { cryptoSource = ""; cryptoIdentity = ""; cryptoPassphrase = "" },
			args:       []string{},
			wantErrMsg: "-s 指定要解密",
		},
		{
			name:       "缺少 -i 和 -p",
			setupVars:  func() { cryptoSource = "/tmp/a.age"; cryptoIdentity = ""; cryptoPassphrase = "" },
			args:       []string{},
			wantErrMsg: "-i 指定私钥文件 或 -p 指定密码",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.setupVars()
			err := runCryptoDecrypt(nil, c.args)
			if err == nil {
				t.Fatalf("期望报错包含 %q, 实际无错误", c.wantErrMsg)
			}
			if !strings.Contains(err.Error(), c.wantErrMsg) {
				t.Errorf("错误信息应该包含 %q, 实际: %v", c.wantErrMsg, err)
			}
		})
	}
}

func TestCryptoKeygen_Execute(t *testing.T) {
	dir := t.TempDir()
	keyPath := dir + "/key.txt"

	cryptoIdentity = keyPath
	err := runCryptoKeygen(nil, []string{})
	if err != nil {
		t.Fatalf("keygen error: %v", err)
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("私钥文件未生成: %v", err)
	}
	if !strings.Contains(string(data), "AGE-SECRET-KEY") {
		t.Errorf("私钥内容不对: %s", data)
	}
}

func TestAllSubcommandsRegistered(t *testing.T) {
	var outBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&outBuf)
	rootCmd.SetArgs([]string{"--help"})
	rootCmd.Execute()

	expected := []string{"crypto", "copy", "move", "rename", "splitzip", "unsplitzip", "icongen", "uuid"}
	for _, name := range expected {
		if !strings.Contains(outBuf.String(), name) {
			t.Errorf("help 输出应该包含子命令: %s", name)
		}
	}
}
