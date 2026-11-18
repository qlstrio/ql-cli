package crypto

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
)

func generateHybridIdentity() (*age.HybridIdentity, error) {
	return age.GenerateHybridIdentity()
}

func encryptFile(inPath, outPath, recipientStr, passphrase string) error {
	in, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	var recipients []age.Recipient
	if recipientStr != "" {
		r, err := age.ParseHybridRecipient(recipientStr)
		if err != nil {
			return fmt.Errorf("解析公钥失败: %w", err)
		}
		recipients = append(recipients, r)
	}
	if passphrase != "" {
		sr, err := age.NewScryptRecipient(passphrase)
		if err != nil {
			return fmt.Errorf("创建密码收件人失败: %w", err)
		}
		recipients = append(recipients, sr)
	}

	w, err := age.Encrypt(out, recipients...)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, in); err != nil {
		return err
	}
	return w.Close()
}

func decryptFile(inPath, outPath, identityPath, passphrase string) error {
	in, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer in.Close()

	var identities []age.Identity
	if identityPath != "" {
		data, err := os.ReadFile(identityPath)
		if err != nil {
			return fmt.Errorf("读取私钥文件失败: %w", err)
		}
		ids, err := age.ParseIdentities(strings.NewReader(string(data)))
		if err != nil {
			return fmt.Errorf("解析私钥失败: %w", err)
		}
		identities = append(identities, ids...)
	}
	if passphrase != "" {
		si, err := age.NewScryptIdentity(passphrase)
		if err != nil {
			return fmt.Errorf("创建密码身份失败: %w", err)
		}
		identities = append(identities, si)
	}

	r, err := age.Decrypt(in, identities...)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, r)
	return err
}
