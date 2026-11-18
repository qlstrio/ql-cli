package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/qlstrio/ql-cli/internal/fileops"
	"github.com/spf13/cobra"
)

var (
	renameDir       string
	renameRecursive bool
	renameFilter    []string
	renameDryRun    bool

	renameFrom      string
	renameTo        string
	renameUseRegex  bool
	renamePrefix    string
	renameSuffix    string
	renameToLower   bool
	renameToUpper   bool
	renameClean     bool
	renameRemoveExt bool
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: color.CyanString("批量修改文件名"),
	Long: color.CyanString(`遍历指定目录,批量修改文件名。支持查找替换、正则匹配、添加前后缀、大小写转换等操作。

示例:
  ql-cli rename -d ./files --from "IMG_" --to "Photo_"
  ql-cli rename -d ./files --regex --from "IMG_(\d+)" --to "Photo_$1"
  ql-cli rename -d ./files --prefix "2024_" --dry-run
  ql-cli rename -d ./files --to-lower --filter ".JPG"`),
	RunE: runRename,
}

func init() {
	renameCmd.Flags().StringVarP(&renameDir, "dir", "d", ".", "要处理的目录路径")
	renameCmd.Flags().BoolVarP(&renameRecursive, "recursive", "r", false, "是否递归处理子目录")
	renameCmd.Flags().BoolVar(&renameDryRun, "dry-run", false, "预览模式(不会实际修改)")
	renameCmd.Flags().StringSliceVarP(&renameFilter, "filter", "f", nil, "筛选指定文件类型")

	renameCmd.Flags().StringVar(&renameFrom, "from", "", "要查找的文本")
	renameCmd.Flags().StringVar(&renameTo, "to", "", "替换为的文本")
	renameCmd.Flags().BoolVar(&renameUseRegex, "regex", true, "是否使用正则表达式匹配")
	renameCmd.Flags().StringVar(&renamePrefix, "prefix", "", "添加文件名前缀")
	renameCmd.Flags().StringVar(&renameSuffix, "suffix", "", "添加文件名后缀(在扩展名前)")
	renameCmd.Flags().BoolVar(&renameToLower, "to-lower", false, "文件名转小写")
	renameCmd.Flags().BoolVar(&renameToUpper, "to-upper", false, "文件名转大写")
	renameCmd.Flags().BoolVar(&renameClean, "clean", false, "清理特殊字符(只保留英文字母、数字、下划线、连字符和中文汉字以及括号)")
	renameCmd.Flags().BoolVar(&renameRemoveExt, "remove-ext", false, "移除扩展名")

	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	opts := fileops.RenameOptions{
		Dir:       renameDir,
		Recursive: renameRecursive,
		Filter:    renameFilter,
		DryRun:    renameDryRun,

		From:      renameFrom,
		To:        renameTo,
		UseRegex:  renameUseRegex,
		Prefix:    renamePrefix,
		Suffix:    renameSuffix,
		ToLower:   renameToLower,
		ToUpper:   renameToUpper,
		Clean:     renameClean,
		RemoveExt: renameRemoveExt,
	}
	if err := opts.ValidateRegex(); err != nil {
		return fmt.Errorf("%s", color.RedString(err.Error()))
	}
	return opts.Rename()
}
