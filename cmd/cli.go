/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package cmd

import (
	"cnabtool/pkg/config"
	"cnabtool/pkg/logging"
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
		Short: "The confluence tool",
		Long: `The tool for manipulating confluence content.

Support verbs get, put and update`,

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			ret := cnf.InitConfig(cmd)
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
	contentCmd.AddCommand(GetContentCmd(cnf))

	// command verb "get" for "content"
	contentCmd.AddCommand(InspectContentCmd(cnf))

	// command verb "get" for "content"
	contentCmd.AddCommand(DeleteContentCmd(cnf))

	return rootCmd
}
