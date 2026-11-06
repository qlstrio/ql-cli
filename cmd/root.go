package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "ql-cli",
	Short:   color.CyanString("CLI工具集"),
	Long:    color.CyanString("一个ql-cli工具集合"),
	Version: color.CyanString("0.0.3"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			showBanner()
		}
		err := cmd.Help()
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func showBanner() {
	banner := `
 ┌───────────────┐
 │     /\_/\     │
 │    ( o.o )    │
 │     > ^ <     │
 │               │
 │   >  ql-cli   │
 └───────────────┘
`
	fmt.Println(color.GreenString(banner))
}
