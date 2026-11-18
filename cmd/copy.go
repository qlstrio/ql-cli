package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/qlstrio/ql-cli/internal/fileops"
	"github.com/spf13/cobra"
)

var (
	copyDir       string
	copyOut       string
	copyRecursive bool
	copyFilter    []string
	copyKeepDir   bool
	copyDryRun    bool
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: color.CyanString("复制文件到指定目录"),
	Long:  color.CyanString("复制文件到指定目录，支持筛选指定文件类型、递归，保留文件结构。"),
	RunE:  runCopy,
}

func init() {
	copyCmd.Flags().StringVarP(&copyDir, "dir", "d", ".", "源目录路径")
	copyCmd.Flags().BoolVarP(&copyRecursive, "recursive", "r", false, "是否递归处理子目录")
	copyCmd.Flags().BoolVar(&copyDryRun, "dry-run", false, "预览模式(不会实际复制)")
	copyCmd.Flags().StringSliceVarP(&copyFilter, "filter", "f", nil, "筛选指定文件类型,支持多种-f=png -f=txt，也支持逗号 -f=png,txt")
	copyCmd.Flags().StringVarP(&copyOut, "out", "o", ".", "输出目录")
	copyCmd.Flags().BoolVarP(&copyKeepDir, "keep-dir", "k", false, "保留目录结构")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	opts := fileops.CommonOptions{
		Dir:       copyDir,
		OutDir:    copyOut,
		Recursive: copyRecursive,
		Filter:    copyFilter,
		KeepDir:   copyKeepDir,
		DryRun:    copyDryRun,
	}
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("%s", color.RedString(err.Error()))
	}
	return fileops.Copy(opts)
}
