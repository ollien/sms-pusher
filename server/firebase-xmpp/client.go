package firebase_xmpp

import "encoding/json"
import "fmt"
import "github.com/mattn/go-xmpp"
import "log"
import "os"


const FCM_SERVER = "fcm-xmpp.googleapis.com"
const FCM_DEV_PORT = 5236
const FCM_PROD_PORT = 5235
const FCM_USERNAME_ADDRESS = "gcm.googleapis.com"

type FirebaseClient struct {
	xmppClient xmpp.Client
	senderId string
	serverKey string
}

type Config struct {
	SenderId string
	ServerKey string
}

//Will create a FirebaseClient from configuration file
func NewFirebaseClient(configPath string) FirebaseClient {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	jsonDecoder := json.NewDecoder(file)
	var config Config
	jsonDecoder.Decode(&config)
	username = fmt.Sprintf("%s@%s", config.ServerKey, FCM_USERNAME_ADDRESS)
	client, err := xmpp.NewClient(FCM_SERVER, username, config.ServerKey, true)
	if err != nil {
		log.Fatal(err)
	}
	return FirebaseClient{
		xmppClient: *client,
		senderId: config.SenderId,
		serverKey: config.ServerKey,
	}
}
