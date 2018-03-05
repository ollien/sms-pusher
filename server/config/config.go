package config

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

const configPath = "config.go"

//Config represents the config for the application
type Config struct {
	Database databaseConfig `json:"db"`
	XMPP     xmppConfig     `json:"xmpp"`
}

//databaseConfig represents the config for the database
type databaseConfig struct {
	URI string `json:"uri"`
}

//xmppConfig represents the config for the XMPP server
type xmppConfig struct {
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
	var config Config
	decoder.Decode(&config)

	return config
}
