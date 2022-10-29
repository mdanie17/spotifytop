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
	// ServerHost specifies the host the server listens on
	ServerHost string `mapstructure:"server_host"`
	// ServerPort specifies the port the server listens on
	ServerPort string `mapstructure:"server_port"`
	// CookieKey specifies the key used for encoding the cookies
	Cookiekey string `mapstructure:"cookie_key"`

	// SpotifyState specifies the string spotify uses to generate unique URLs
	SpotifyState string `mapstructure:"spotify_state"`
	// SpotifyRedirectURI specifies the URI that will be redirected
	// to from spotify upon succesful login
	// This needs to include protocol and port, e.g:
	// http://localhost:8080
	SpotifyRedirectURI string `mapstructure:"spotify_redirect_uri"`
	// SpotifyClientKey is the client key specified on the spotify
	// developer portal
	SpotifyClientKey string `mapstructure:"spotify_client_key"`
	// SpotifySecretKey is the client key specified on the spotify
	// developer portal
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
