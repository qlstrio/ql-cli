package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/qlstrio/ql-cli/internal/splitzip"
	"github.com/spf13/cobra"
)

var splitZipCmd = &cobra.Command{
	Use:   "splitzip",
	Short: color.CyanString("将文件分割压缩为多个zip文件"),
	Long:  color.CyanString("将文件分割压缩为多个zip文件，保留文件结构。"),
	RunE:  runSplitZip,
}
var (
	zipFolder    string
	outputFolder string
	zPrefix      string
	maxSize      int64
)

func init() {
	splitZipCmd.Flags().StringVarP(&zipFolder, "source", "s", "", "源文件目录")
	splitZipCmd.Flags().StringVarP(&outputFolder, "dir", "d", ".", "输出目录")
	splitZipCmd.Flags().StringVarP(&zPrefix, "prefix", "p", "split_", "zip文件前缀")
	splitZipCmd.Flags().Int64VarP(&maxSize, "max", "b", 2, "最大文件大小，单位MB，若单文件大小超过最大文件大小，单独压缩")
	rootCmd.AddCommand(splitZipCmd)
}

func runSplitZip(cmd *cobra.Command, args []string) error {
	var err error
	zipFolder, err = getAbsDirPath(zipFolder)
	if err != nil {
		return err
	}
	outputFolder, err = getAbsDirPath(outputFolder)
	if err != nil {
		return err
	}
	splitZip.DoZip(zipFolder, outputFolder, maxSize*1024*1024, zPrefix)
	return nil
}

var unSplitZipCmd = &cobra.Command{
	Use:   "unsplitzip",
	Short: color.CyanString("将多个分割的zip文件还原为原始目录"),
	Long:  color.CyanString("将 splitzip 生成的多个 zip 文件解压到同一目录，自动合并还原文件结构。"),
	RunE:  runUnsplitZip,
}

var (
	unzipSource string
	unzipOutput string
	unzipPrefix string
)

func init() {
	unSplitZipCmd.Flags().StringVarP(&unzipSource, "source", "s", "", "包含分割zip文件的目录")
	unSplitZipCmd.Flags().StringVarP(&unzipOutput, "dir", "d", ".", "还原输出目录")
	unSplitZipCmd.Flags().StringVarP(&unzipPrefix, "prefix", "p", "split_", "zip文件前缀(与splitzip一致)")
	rootCmd.AddCommand(unSplitZipCmd)
}

func runUnsplitZip(cmd *cobra.Command, args []string) error {
	var err error
	unzipSource, err = getAbsDirPath(unzipSource)
	if err != nil {
		return err
	}
	unzipOutput, err = filepath.Abs(unzipOutput)
	if err != nil {
		return fmt.Errorf(color.RedString("获取输出目录绝对路径失败: %w"), err)
	}

	pattern := filepath.Join(unzipSource, unzipPrefix+"*.zip")
	zipFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf(color.RedString("查找zip文件失败: %w"), err)
	}
	if len(zipFiles) == 0 {
		return fmt.Errorf(color.RedString("在 %s 下未找到匹配 %s*.zip 的文件"), unzipSource, unzipPrefix)
	}

	return splitZip.DoExtract(zipFiles, unzipOutput)
}
