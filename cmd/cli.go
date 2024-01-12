/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package cmd

import (
	"cnabtool/pkg/config"
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"encoding/base64"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strconv"
)

var Version string
var Commit string

func BuildCliCmd(cnf *config.Config) *cobra.Command {

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "cnabtool",
		Short: "The cnab tool",
		Long: `The tool for manipulating cnab content.
`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			ret := cnf.InitConfig(cmd)

			// add sensitives to global list
			data.Sensitives = append(data.Sensitives, data.Gc.Credentials.Password)
			basicauth := []byte(data.Gc.Credentials.Username + ":" + data.Gc.Credentials.Password)
			data.Sensitives = append(data.Sensitives, base64.StdEncoding.EncodeToString(basicauth))

			return ret
		},
	}

	// config file path
	rootCmd.PersistentFlags().StringP("config", "c", "", "Customer config file path.")

	// common flags
	rootCmd.PersistentFlags().IntVarP(&cnf.Verbosity, "verbosity", "v", config.ConfigDefaultVerbosity,
		"Log level verbosity. For quiet use "+strconv.Itoa(logging.LogQuietLevel)+".")
	viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))

	rootCmd.PersistentFlags().StringVarP(&cnf.Credentials.Username, "username", "u", "", "Registry User name.")
	viper.BindPFlag("server.username", rootCmd.PersistentFlags().Lookup("username"))

	rootCmd.PersistentFlags().StringVarP(&cnf.Credentials.Password, "password", "p", "", "Registry password.")
	viper.BindPFlag("server.token", rootCmd.PersistentFlags().Lookup("token"))

	rootCmd.PersistentFlags().IntVarP(&cnf.Timeout, "url", "t", 10000, "Timeout ms.")
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

	rootCmd.AddCommand(VersionCmd(cnf))

	// command noun "content"
	contentCmd := ContentCmd(cnf)
	rootCmd.AddCommand(contentCmd)

	// command verb "get" for "content"
	contentCmd.AddCommand(GetManifestCmd(cnf))

	// command verb "inspect" for "content"
	inspectContentCmd := InspectContentCmd(cnf)
	contentCmd.AddCommand(inspectContentCmd)
	// local flag raw outputs
	inspectContentCmd.Flags().BoolVarP(&cnf.Raw, "raw", "", false, "Raw format for inspected content")

	// command verb "delete" for "content"
	deleteContentCmd := DeleteContentCmd(cnf)
	contentCmd.AddCommand(deleteContentCmd)
	// local flag dry-run
	deleteContentCmd.Flags().BoolVarP(&cnf.DryRun, "dry-run", "", false, "Dry-run mode")

	return rootCmd
}
