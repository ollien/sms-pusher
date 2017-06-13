package firebase_xmpp

import "encoding/json"
import "github.com/processone/gox/xmpp"
import "log"
import "os"

type FirebaseClient struct {
	xmppClient *xmpp.Client
	senderId string
	serverKey string
}

type Config struct {
	SenderId string
	ServerKey string
}

//Will create a FirebaseClient from configuration file
func NewFirebaseClient(configPath string) FirebaseClient {
	file, error := os.Open(configPath)
	if error != nil {
		log.Fatal(error)
	}
	jsonDecoder := json.NewDecoder(file)
	var config Config
	jsonDecoder.Decode(&config)
	return FirebaseClient{
		senderId: config.SenderId,
		serverKey: config.ServerKey,
	}
}
