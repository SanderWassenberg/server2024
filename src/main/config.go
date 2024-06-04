package main

import (
	"errors"
	"fmt"
	"os"

	gc "src/generic_config"
)

type Config struct {
	SendGrid_api_key string
	Port             string
}

var config Config

func load_config_or_exit() {
	const filename = "secret.config"

	if err := gc.LoadConfig(filename, &config); err != nil {

		fmt.Printf("Error(s) loading config.\n%v\n", err)

		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf(
`Missing config.
To add a config, make a new file "%v" and place it next to this executable.
Files with prefix "secret." won't be checked in to version control.

The config format is very simple:

	# Comment. This only works on its own line, don't place them after values
	Key:      Value
	DeadBeef: 3735928559

The type associated with each key (string, int etc.) is defined in the program.
You'll be informed of any values that are missing from the config upon startup.
`, filename)
		}
		os.Exit(1)
	}
}

