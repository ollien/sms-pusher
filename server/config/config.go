package config

//Config represents the config for the application
type Config struct {
	Database databaseConfig
	XMPP     xmppConfig
}

//databaseConfig represents the config for the database
type databaseConfig struct {
	URI string
}

//xmppConfig represents the config for the XMPP server
type xmppConfig struct {
	ServerKey string
	SenderID  string
}
