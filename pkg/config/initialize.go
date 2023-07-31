package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("error while reading config file: %s", err.Error())
	}
	for _, env := range viper.AllKeys() {
		if viper.GetString(env) != "" {
			_ = os.Setenv(env, viper.GetString(env))
			_ = os.Setenv(strings.ToUpper(env), viper.GetString(env))
		}
	}
}
