package config

import (
	"github.com/spf13/viper"
)

// Registry is for the configuration values.
var Registry *viper.Viper

// Set the configs
func Set() {
	v := viper.New()

	v.AddConfigPath(".")
	v.AddConfigPath("../..")

	v.AutomaticEnv()

	Registry = v
}
