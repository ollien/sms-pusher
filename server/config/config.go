package config

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
