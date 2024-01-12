/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package cmd

import (
	"cnabtool/pkg/client"
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
			logging.Fatal("too a few arguments. use action's verb")
		},
	}

	return contentCmd
}

// GetManifestCmd get a content index from registry

func GetManifestCmd(cnf *config.Config) *cobra.Command {

	// cmd represents the content command
	var getContentCmd = &cobra.Command{
		Use:   "manifest",
		Short: "Get the content manifest",
		Long:  `Get manifest with reference address and show it as json`,

		Run: func(cc *cobra.Command, args []string) {
			// argument is registry reference string
			if len(args) == 0 {
				logging.Fatal("too a few arguments. use reference to index")
			}

			config := (*content.Config)(cnf)

			logging.Debug(fmt.Sprintf("config %+v", config))
			response, _, err := config.GetManifest(args[0])
			if err == nil {
				if data.Gc.Verbosity >= logging.LogNormalLevel {
					content.ResponsePrettyPrint(response)
				}
			} else {
				logging.Error(fmt.Sprintf("%+v", err))
			}
		},
	}

	return getContentCmd
}

// InspectContentCmd inspect content of the cnab project

func InspectContentCmd(cnf *config.Config) *cobra.Command {

	// cmd represents the content command
	var inspectContentCmd = &cobra.Command{
		Use:   "inspect",
		Short: "Inspect the content",
		Long: `Inspect all items in project with manifest reference address
and report summary as json`,

		Run: func(cc *cobra.Command, args []string) {
			if len(args) == 0 {
				logging.Fatal("too a few arguments. use reference to cnab")
			}

			config := (*content.Config)(cnf)

			logging.Debug(fmt.Sprintf("config %+v", config))
			regres, cl, err := config.GetManifest(args[0])
			if err != nil {
				logging.Error(fmt.Sprintf("%+v", err))
			} else {
				if regres.Media != client.MediaTypeOciIndex {
					logging.Error(fmt.Sprintf("unexpected media type %+v, must be cnab index", regres.Media))
					// add comment to error
					errLine, err := logging.PrettyString(regres.Content)
					if err != nil {
						errLine = regres.Content
					}
					logging.Error(fmt.Sprintf("%+v", errLine))
				} else {
					// continue if first reference is cnab index
					// add first index
					err := content.AddCnab(regres, cl.Tag)
					if err != nil {
						errLine := fmt.Sprintf("can't create first index, %+v", err.Error())
						logging.Fatal(errLine)
					}
					config.InspectCnab(cl)
					config.ShowCnabReport(cl)
				}
			}
		},
	}

	return inspectContentCmd
}

// DeleteContentCmd delete all possible parts of the cnab components

func DeleteContentCmd(cnf *config.Config) *cobra.Command {

	// cmd represents the content command
	var deleteContentCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete the cnab content",
		Long: `Inspect cnab project and delete all possible component parts of
selected cnab`,

		Run: func(cc *cobra.Command, args []string) {
			if len(args) == 0 {
				logging.Fatal("too a few arguments. use reference to cnab")
			}

			config := (*content.Config)(cnf)

			logging.Debug(fmt.Sprintf("config %+v", config))
			regres, cl, err := config.GetManifest(args[0])
			if err != nil {
				logging.Error(fmt.Sprintf("%+v", err))
			} else {
				if regres.Media != client.MediaTypeOciIndex {
					logging.Error(fmt.Sprintf("unexpected media type %+v, must be cnab index", regres.Media))
					// add comment to error
					errLine, err := logging.PrettyString(regres.Content)
					if err != nil {
						errLine = regres.Content
					}
					logging.Error(fmt.Sprintf("%+v", errLine))
				} else {
					// continue if first reference is cnab index
					// add first index
					err := content.AddCnab(regres, cl.Tag)
					if err != nil {
						errLine := fmt.Sprintf("can't create first index, %+v", err.Error())
						logging.Fatal(errLine)
					}
					config.InspectCnab(cl)
					if data.Gc.Verbosity >= logging.LogDebugLevel {
						config.ShowCnabReport(cl)
					}
					config.DeleteCnab(cl)
				}
			}
		},
	}

	return deleteContentCmd
}
