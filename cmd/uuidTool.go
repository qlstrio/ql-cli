package cmd

import (
	"github.com/fatih/color"
	"github.com/qlstrio/ql-cli/cmd/uuidTool"
	"github.com/spf13/cobra"
)

var uuidToolCmd = &cobra.Command{
	Use:   "uuid",
	Short: color.CyanString("UUID 工具"),
	Long:  color.CyanString("UUID 工具"),
	RunE:  runUUIDTool,
}

func init() {
	uuidToolCmd.AddCommand(uuidTool.GenerateCmd, uuidTool.CompressCmd, uuidTool.DecompressCmd)
	rootCmd.AddCommand(uuidToolCmd)
}

func runUUIDTool(cmd *cobra.Command, args []string) error {
	err := cmd.Help()
	if err != nil {
		return err
	}
	return nil
}
