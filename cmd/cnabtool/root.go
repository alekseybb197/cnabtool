package cnabtool

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:     "cnabtool",
	Version: version,
	Short:   "cnabtool - a cnab cli tool",
	Long: `cnabtool is a cnab cli tool
One can use it to manipulate cnab artifacts`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error '%s'", err)
		os.Exit(1)
	}
}
