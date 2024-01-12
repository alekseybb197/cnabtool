/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package config

import (
	"cnabtool/pkg/data"
	"cnabtool/pkg/logging"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

const (
	ConfigFileName         = "config"
	ConfigFileExt          = "yaml"
	ConfigFileDir          = "cnabtool"
	ConfigEnvPrefix        = "CNAB"
	ConfigDefaultVerbosity = logging.LogNormalLevel
	ConfigDefaultTimeout   = 10000
	ConfigDefaultClient    = "curl/7.79.1"
	ConfigDefaultScheme    = "https"
)

type Config data.Config

// make new default config

func New() *Config {
	if data.Gc == nil {
		cnf := &Config{}
		cnf.Verbosity = ConfigDefaultVerbosity // set default
		cnf.Timeout = ConfigDefaultTimeout
		cnf.Client = ConfigDefaultClient
		cnf.Unsecure = false
		cnf.Raw = false
		cnf.Scheme = ConfigDefaultScheme
		data.Gc = (*data.Config)(cnf)
	}
	return (*Config)(data.Gc)
}

// Initial config from file, environment and etc.

func (cnf *Config) InitConfig(cmd *cobra.Command) error {

	// try apply custom config
	customconfig := cmd.Flags().Lookup("config").Value.String()
	if len(customconfig) != 0 {
		logging.Info("use custom config "+customconfig)
		file_extension := filepath.Ext(customconfig)
		viper.SetConfigName(strings.TrimSuffix(filepath.Base(customconfig), file_extension))
		viper.SetConfigType(strings.TrimPrefix(file_extension, "."))
		viper.AddConfigPath(filepath.Dir(customconfig))
	} else {
		// viper defaults
		viper.SetConfigName(ConfigFileName)                // name of config file (without extension)
		viper.SetConfigType(ConfigFileExt)                 // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath("/etc/" + ConfigFileDir + "/") // path to look for the config file in
		viper.AddConfigPath("$HOME/." + ConfigFileDir)     // call multiple times to add many search paths
		viper.AddConfigPath(".")                           // optionally look for config in the working directory
	}

	// fetch config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			logging.Error(fmt.Sprintf("fatal error config file: %s", err.Error()))
			return err
		} else {
			if len(customconfig) != 0 {
				logging.Fatal("can not found custom config file "+customconfig)
			}
		}
	}

	// fetch environment variables
	viper.SetEnvPrefix(ConfigEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// fetch root level values
	if err := viper.Unmarshal(cnf); err != nil {
		logging.Error(fmt.Sprintf("unable to decode into config struct, %s", err.Error()))
		return err
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name

		if !f.Changed && viper.IsSet(configName) { //
			val := viper.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

	return nil
}
