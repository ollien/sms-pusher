package config

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

const configPath = "config.json"

//Config represents the config for the application
type Config struct {
	Database DatabaseConfig `json:"db"`
	XMPP     XMPPConfig     `json:"xmpp"`
	MMS      MMSConfig      `json:"mms"`
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

//MMSConfig represents the config for the MMS portion of the FCM XMPP server
type MMSConfig struct {
	UploadLocation string `json:"upload_location"`
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
