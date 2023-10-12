package cnabtool

import (
	"fmt"
	"github.com/alekseybb197/cnabtool/pkg/cnabtool"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get request",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		res := cnabtool.Get(args[0])
		fmt.Println(res)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
