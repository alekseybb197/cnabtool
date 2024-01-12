/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package cmd

import (
	"cnabtool/pkg/config"
	"fmt"
	"github.com/spf13/cobra"
)

// contentCmd represents the content command

func VersionCmd(cnf *config.Config) *cobra.Command {

	// cc represents the content command
	var contentCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the application version",

		Run: func(cc *cobra.Command, args []string) {
			fmt.Printf("Version: %s (%s)\n", Version, Commit)
		},
	}

	return contentCmd
}
