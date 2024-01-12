/*
Copyright © 2023 Aleksey Barabanov <alekseybb@gmail.comS>
*/

package main

import (
	"cnabtool/cmd"
	"cnabtool/pkg/config"
	"cnabtool/pkg/data"
	"os"
)

// main procedure
func main() {

	// fix version if not exists
	if len(cmd.Version) == 0 {
		cmd.Version = "0.0.0"
		cmd.Commit = "n/a"
	}

	// make config with defaults and fill values from configs
	cnf := config.New()

	cli := cmd.BuildCliCmd(cnf)

	err := cli.Execute()

	if err != nil || data.Gc.Error > 0 {
		os.Exit(1)
	}
}
