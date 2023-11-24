/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package data

// config structure

type Config struct {
	// configuration
	Verbosity int    `mapstructure:"verbosity"` // log level
	Timeout   int    `mapstructure:"timeout"`   // web io timeout ms
	Unsecure  bool   `mapstructure:"unsecure"`  // unsecure tls
	Client    string `mapstructure:"client"`    // http client
	Scheme    string `mapstructure:"scheme"`    // url scheme
	// credentials
	Credentials Credentials `mapstructure:"credentials"`

	//WebClient http.Client // web client
}

// server url and credentials

type Credentials struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// global config

var Gc *Config = nil
