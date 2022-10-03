package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	SoftwareName = "spotifytop"
)

type ServerConfig struct {
	ServerHost string `mapstructure:"server_host"`
	ServerPort string `mapstructure:"server_port"`
	State      string `mapstructure:"spotify_state"`
	Cookiekey  string `mapstructure:"cookie_key"`

	SpotifyClientKey string `mapstructure:"spotify_client_key"`
	SpotifySecretKey string `mapstructure:"spotify_secret_key"`
}

// Generates the path string
func userConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get user config dir")
	}

	return fmt.Sprintf("%s/%s", dir, strings.ToLower(SoftwareName))
}

func GetServerConfig() (ServerConfig, error) {
	vip := viper.New()
	log.Info().Str("config_path", userConfigDir()).Msg("setting config path")

	// setup viper
	vip.SetConfigName("config")
	vip.SetConfigType("yaml")
	vip.AddConfigPath(userConfigDir())
	vip.SetEnvPrefix(strings.ToUpper(SoftwareName))
	vip.AutomaticEnv()
	setClientDefaults(vip)

	// read configuration file
	// if the file exists and malformatted, panic.
	// it it does not exists, just continue
	if err := vip.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return ServerConfig{}, err
		}
	}

	// unrmarshal configuration into struct
	var conf ServerConfig
	if err := vip.Unmarshal(&conf); err != nil {
		return ServerConfig{}, err
	}

	return conf, nil
}

func setClientDefaults(vip *viper.Viper) {
	// set default values
	vip.SetDefault("spotify_state", "secret")
	vip.SetDefault("cookie_key", "secret")
}
