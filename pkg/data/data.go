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
	Raw       bool   `mapstructure:"raw"`       // raw format - only for inspect content
	DryRun    bool   `mapstructure:"dryrun"`    // dry-run mode - only for delete content
	Error     int    // errors count
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

// list fo masked literals

var Sensitives []string

// link to project root

var Scheme string
var Registry string
var Repository string

// items queue

var ItemsQueue []string

// cnab index item

type CnabItem struct {
	Digest     string
	Annotation string
}

// cnab catalog

type RegIndex struct {
	Reference  string
	Tag        string
	Media      string
	Annotation string
	Date       string
	Digest     string
	DownLinks  []CnabItem // down links
	UpLinks    []CnabItem // up links
	Lost       int        // link not found
	Content    string     // pretty json
}

var ProjectList []*RegIndex

var ItemByDigest = make(map[string]*RegIndex)
var ItemByTag = make(map[string]*RegIndex)

// catalog item types

const (
	ItemTypeCnab   = "cnab index"
	ItemTypeImage  = "docker image"
	ItemTypeConfig = "cnab config"
	ItemTypeStuff  = "stuff"
)
