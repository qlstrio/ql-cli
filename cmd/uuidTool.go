package cmd

import (
	"fmt"

	"github.com/qlstrio/ql-cli/cmd/uuidTool"
	"github.com/spf13/cobra"
)

var uuidToolCmd = &cobra.Command{
	Use:   "uuid",
	Short: "UUID 工具",
	Long:  "UUID 工具",
	RunE:  runUUIDTool,
}

func init() {
	uuidToolCmd.AddCommand(uuidTool.GenerateCmd, uuidTool.CompressCmd, uuidTool.DecompressCmd)
	rootCmd.AddCommand(uuidToolCmd)
}

func runUUIDTool(cmd *cobra.Command, args []string) error {
	fmt.Println("UUID 工具")
	return nil
}
