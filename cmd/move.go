package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/qlstrio/ql-cli/internal/fileops"
	"github.com/spf13/cobra"
)

var (
	moveDir       string
	moveOut       string
	moveRecursive bool
	moveFilter    []string
	moveKeepDir   bool
	moveDryRun    bool
)

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: color.CyanString("移动文件到指定目录"),
	Long:  color.CyanString("移动文件到指定目录，支持筛选指定文件类型、递归，保留文件结构。"),
	RunE:  runMove,
}

func init() {
	moveCmd.Flags().StringVarP(&moveDir, "dir", "d", ".", "源目录路径")
	moveCmd.Flags().BoolVarP(&moveRecursive, "recursive", "r", false, "是否递归处理子目录")
	moveCmd.Flags().BoolVar(&moveDryRun, "dry-run", false, "预览模式(不会实际移动)")
	moveCmd.Flags().StringSliceVarP(&moveFilter, "filter", "f", nil, "筛选指定文件类型,支持多种-f=png -f=txt，也支持逗号 -f=png,txt")
	moveCmd.Flags().StringVarP(&moveOut, "out", "o", ".", "输出目录")
	moveCmd.Flags().BoolVarP(&moveKeepDir, "keep-dir", "k", false, "保留目录结构")
	rootCmd.AddCommand(moveCmd)
}

func runMove(cmd *cobra.Command, args []string) error {
	opts := fileops.CommonOptions{
		Dir:       moveDir,
		OutDir:    moveOut,
		Recursive: moveRecursive,
		Filter:    moveFilter,
		KeepDir:   moveKeepDir,
		DryRun:    moveDryRun,
	}
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("%s", color.RedString(err.Error()))
	}
	return fileops.Move(opts)
}
