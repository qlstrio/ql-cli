package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/qlstrio/ql-cli/internal/crypto"
)

var (
	cryptoSource     string
	cryptoRecipient  string
	cryptoIdentity   string
	cryptoPassphrase string
	cryptoOut        string
	cryptoRecursive  bool
	cryptoFilter     []string
	cryptoKeepDir    bool
	cryptoSuffix     string
)

var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: color.CyanString("文件加解密工具 (基于 age)"),
	Long: color.CyanString(`支持两种加密方式:
  1. 密码短语   -p/--passphrase
  2. 公钥/私钥   -r/--recipient (加密) + -i/--identity (解密)

支持: 单文件 / 目录 / 递归目录 / 按后缀筛选
公钥格式: age1...
密钥对生成: ql-cli crypto keygen [-o key.txt]`),
}

var cryptoEncryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "加密文件/目录",
	RunE:  runCryptoEncrypt,
}

var cryptoDecryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "解密文件/目录",
	RunE:  runCryptoDecrypt,
}

var cryptoKeygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "生成 age 密钥对",
	RunE:  runCryptoKeygen,
}

func init() {
	cryptoEncryptCmd.Flags().StringVarP(&cryptoSource, "source", "s", "", "要加密的文件或目录路径")
	cryptoEncryptCmd.Flags().StringVarP(&cryptoRecipient, "recipient", "r", "", "age 公钥 (age1...)")
	cryptoEncryptCmd.Flags().StringVarP(&cryptoPassphrase, "passphrase", "p", "", "密码短语")
	cryptoEncryptCmd.Flags().StringVarP(&cryptoOut, "out", "o", "", "输出路径")
	cryptoEncryptCmd.Flags().BoolVarP(&cryptoRecursive, "recursive", "R", false, "递归处理子目录")
	cryptoEncryptCmd.Flags().StringSliceVarP(&cryptoFilter, "filter", "f", nil, "按后缀筛选, 例 -f=txt -f=json (默认全部)")
	cryptoEncryptCmd.Flags().BoolVarP(&cryptoKeepDir, "keep-dir", "k", false, "目录模式下保留子目录结构")
	cryptoEncryptCmd.Flags().StringVar(&cryptoSuffix, "suffix", "", "加密时追加的后缀, 例 --suffix=.age")

	cryptoDecryptCmd.Flags().StringVarP(&cryptoSource, "source", "s", "", "要解密的文件或目录路径")
	cryptoDecryptCmd.Flags().StringVarP(&cryptoIdentity, "identity", "i", "", "age 私钥文件路径")
	cryptoDecryptCmd.Flags().StringVarP(&cryptoPassphrase, "passphrase", "p", "", "密码短语")
	cryptoDecryptCmd.Flags().StringVarP(&cryptoOut, "out", "o", "", "输出路径")
	cryptoDecryptCmd.Flags().BoolVarP(&cryptoRecursive, "recursive", "R", false, "递归处理子目录")
	cryptoDecryptCmd.Flags().StringSliceVarP(&cryptoFilter, "filter", "f", nil, "按后缀筛选 (默认全部)")
	cryptoDecryptCmd.Flags().BoolVarP(&cryptoKeepDir, "keep-dir", "k", false, "目录模式下保留子目录结构")
	cryptoDecryptCmd.Flags().StringVar(&cryptoSuffix, "suffix", "", "解密时要去除的后缀, 例 --suffix=.age")

	cryptoKeygenCmd.Flags().StringVarP(&cryptoIdentity, "out", "o", "", "私钥输出路径 (默认 ~/.age/key.txt)")

	cryptoCmd.AddCommand(cryptoEncryptCmd, cryptoDecryptCmd, cryptoKeygenCmd)
	rootCmd.AddCommand(cryptoCmd)
}

func runCryptoKeygen(cmd *cobra.Command, args []string) error {
	pub, err := crypto.Keygen(cryptoIdentity)
	if err != nil {
		return err
	}
	fmt.Printf("公钥: %s\n", color.GreenString(pub))
	return nil
}

func runCryptoEncrypt(cmd *cobra.Command, args []string) error {
	if cryptoSource == "" {
		return fmt.Errorf("%s", color.RedString("请通过 -s 指定要加密的文件或目录"))
	}
	if cryptoRecipient == "" && cryptoPassphrase == "" {
		return fmt.Errorf("%s", color.RedString("请通过 -r 指定公钥 或 -p 指定密码短语"))
	}

	return crypto.Encrypt(crypto.Options{
		Source:     cryptoSource,
		Out:        cryptoOut,
		Recursive:  cryptoRecursive,
		Filter:     cryptoFilter,
		KeepDir:    cryptoKeepDir,
		Suffix:     cryptoSuffix,
		Recipient:  cryptoRecipient,
		Passphrase: cryptoPassphrase,
	})
}

func runCryptoDecrypt(cmd *cobra.Command, args []string) error {
	if cryptoSource == "" {
		return fmt.Errorf("%s", color.RedString("请通过 -s 指定要解密的文件或目录"))
	}
	if cryptoIdentity == "" && cryptoPassphrase == "" {
		return fmt.Errorf("%s", color.RedString("请通过 -i 指定私钥文件 或 -p 指定密码短语"))
	}

	return crypto.Decrypt(crypto.Options{
		Source:     cryptoSource,
		Out:        cryptoOut,
		Recursive:  cryptoRecursive,
		Filter:     cryptoFilter,
		KeepDir:    cryptoKeepDir,
		Suffix:     cryptoSuffix,
		Identity:   cryptoIdentity,
		Passphrase: cryptoPassphrase,
	})
}
