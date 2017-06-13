package firebase_xmpp

import "encoding/json"
import "fmt"
import "github.com/processone/gox/xmpp"
import "log"
import "os"


const FCM_SERVER = "fcm-xmpp.googleapis.com"
const FCM_DEV_PORT = 5236
const FCM_PROD_PORT = 5235
const FCM_USERNAME_ADDRESS = "gcm.googleapis.com"

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
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	jsonDecoder := json.NewDecoder(file)
	var config Config
	jsonDecoder.Decode(&config)
	clientOptions := generateClientOptions(config)
	client, err := xmpp.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	return FirebaseClient{
		xmppClient: client,
		senderId: config.SenderId,
		serverKey: config.ServerKey,
	}
}

func generateClientOptions(config Config) xmpp.Options {
	//TODO: Check if dev or production, and change FCM port accordingly
	//As a workaround, to prevent go from telling me I have unused variables, I'm saving the production prot to an unused identifier
	//For now, we will always set the port to the dev port
	_ = FCM_PROD_PORT
	port := FCM_DEV_PORT
	return xmpp.Options {
		Address: fmt.Sprintf("%s:%d", FCM_SERVER,port),
		Jid: fmt.Sprintf("%s@%s", config.SenderId, FCM_USERNAME_ADDRESS),
		Password: config.ServerKey,
		Retry: 10,
		ConnectTimeout: 15,
	}
}
