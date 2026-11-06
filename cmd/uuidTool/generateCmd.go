package uuidTool

import (
	"fmt"
	"uuid"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var GenerateCmd = &cobra.Command{
	Use:   "gen",
	Short: "生成 UUID",
	RunE:  runGenerate,
}

func runGenerate(cmd *cobra.Command, args []string) error {
	fmt.Println("生成 UUID")
	fmt.Println(color.YellowString(uuid.New().String()))
	return nil
}
