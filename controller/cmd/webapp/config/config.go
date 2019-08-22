package config

import (
	"fmt"

	"egbitbucket.dtvops.net/com/goatt/pkg/config"
	"github.com/spf13/viper"
)

// Registry is for the configuration values.
var Registry *viper.Viper

// Set the configs
func Set() {
	var err error
	config.AddConfigPath(".")
	config.AddConfigPath("../..")

	err = config.SetConfig(true)
	if err != nil {
		err = fmt.Errorf("failed to set goatt config: %v", err)
		panic(err)
	}
	Registry = config.Viper
}
