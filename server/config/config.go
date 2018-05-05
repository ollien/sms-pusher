package config

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

const configPath = "config.json"

var config Config

//Config represents the config for the application
type Config struct {
	Database DatabaseConfig `json:"db"`
	XMPP     XMPPConfig     `json:"xmpp"`
}

//DatabaseConfig represents the config for the database
type DatabaseConfig struct {
	URI string `json:"uri"`
}

//XMPPConfig represents the config for the XMPP server
type XMPPConfig struct {
	ServerKey string `json:"server_key"`
	SenderID  string `json:"sender_id"`
}

//ParseConfig parses the default configPath into a Config
func ParseConfig(logger *logrus.Logger) Config {
	configFile, err := os.Open(configPath)
	if err != nil {
		logger.Fatalf("Error reading config: %s", err)
	}

	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	var parsedConfig Config
	decoder.Decode(&parsedConfig)

	config = parsedConfig

	return config
}

//GetConfig will get the currently stored config
func GetConfig(logger *logrus.Logger) Config {
	if config == (Config{}) {
		ParseConfig(logger)
	}

	return config
}
