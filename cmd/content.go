/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package cmd

import (
	"cnabtool/pkg/config"
	"cnabtool/pkg/content"
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"fmt"
	"github.com/spf13/cobra"
)

// contentCmd represents the content command

func ContentCmd(cnf *config.Config) *cobra.Command {

	// cc represents the content command
	var contentCmd = &cobra.Command{
		Use:   "content",
		Short: "Content manipulation",
		Long:  `Get indexes, inspect components of registry objects and delete cnab and images`,
		Run: func(cc *cobra.Command, args []string) {
			logging.Fatal("content cmd", "too a few arguments. use action's verb")
		},
	}

	return contentCmd
}

// GetContentCmd represents the content command

func GetContentCmd(cnf *config.Config) *cobra.Command {

	// cmd represents the content command
	var getContentCmd = &cobra.Command{
		Use:   "get",
		Short: "Get the content manifest",
		Long:  `Get manifest with reference address and show it as json`,

		Run: func(cc *cobra.Command, args []string) {
			// argument is registry reference string
			if len(args) == 0 {
				logging.Fatal("get content", "too a few arguments. use document's id")
			}

			config := (*content.Config)(cnf)

			logging.Debug("GetContentCmd", fmt.Sprintf("config %+v", config))
			config.GetManifest(args[0])

		},
	}

	return getContentCmd
}

// GetContentCmd represents the content command

func InspectContentCmd(cnf *config.Config) *cobra.Command {

	// cmd represents the content command
	var inspectContentCmd = &cobra.Command{
		Use:   "inspect",
		Short: "Inspect the content",
		Long: `Inspect all items in project with manifest reference address
and report summary as json`,

		Run: func(cc *cobra.Command, args []string) {
			if len(args) == 0 {
				logging.Fatal("get content", "too a few arguments. use document's id")
			}

			fmt.Println("================= Inspect content called ==============")
			fmt.Printf("current config %+v\n", *cnf)
			fmt.Printf("args %+v\n", args)

			logging.Info("InspectContentCmd", fmt.Sprintf("config %+v", data.Gc))
		},
	}

	return inspectContentCmd
}

// GetContentCmd represents the content command

func DeleteContentCmd(cnf *config.Config) *cobra.Command {

	// cmd represents the content command
	var deleteContentCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete the content",

		Run: func(cc *cobra.Command, args []string) {
			if len(args) == 0 {
				logging.Fatal("delete content", "too a few arguments. use document's id")
			}

			fmt.Println("================= delete content called ==============")
			fmt.Printf("current config %+v\n", *cnf)
			fmt.Printf("args %+v\n", args)
		},
	}

	return deleteContentCmd
}
