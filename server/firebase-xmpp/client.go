package firebasexmpp

import "encoding/json"
import "fmt"
import "github.com/mattn/go-xmpp"
import "log"
import "os"
import "time"


const fcmServer = "fcm-xmpp.googleapis.com"
const fcmDevPort = 5236
const fcmProdPort = 5235
const fcmUsernameAddres = "gcm.googleapis.com"

//FirebaseClient stores the data necessary to be an XMPP Client for Firebase Cloud Messaging. See the spec at https://firebase.google.com/docs/cloud-messaging/xmpp-server-ref
type FirebaseClient struct {
	xmppClient xmpp.Client
	senderID string
	serverKey string
}

//Config stores the details necessary for authenticating to Firebase Cloud Messaging's XMPP server, which cannot be hardcoded or put into version control.
type Config struct {
	SenderID string
	ServerKey string
}

//NewFirebaseClient creates a FirebaseClient from configuration file.
func NewFirebaseClient(configPath string) FirebaseClient {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	jsonDecoder := json.NewDecoder(file)
	var config Config
	jsonDecoder.Decode(&config)
	//TODO: Detect if debug. For now, use dev port and set debug to true. For now, we will just always do this.
	server := fmt.Sprintf("%s:%d", fcmServer, fcmDevPort)
	username := fmt.Sprintf("%s@%s", config.SenderID, fcmUsernameAddres)
	client, err := xmpp.NewClient(server, username, config.ServerKey, true)
	if err != nil {
		log.Fatal(err)
	}
	return FirebaseClient{
		xmppClient: *client,
		senderID: config.SenderID,
		serverKey: config.ServerKey,
	}
}

//recv listens for incomgin messages from Firebase Cloud Messaging.
func (client *FirebaseClient) recv(recvChannel chan<- SMSMessage, drainChannel chan<- ConnectionDrainingMessage) {
	for {
		data, err := client.xmppClient.Recv()
		if err != nil {
			log.Fatal(err)
		}
		chat := data.(xmpp.Chat)
		messageBody := []byte(chat.Other[0])
		messageType, err := GetMessageType(messageBody)
		if err != nil {
			//Don't need to quit for unknowm message types
			log.Println(err)
		} else if (messageType == "UpstreamMessage") {
			var message UpstreamMessage
			json.Unmarshal(messageBody, &message)
			_, err := client.sendACK(message)
			if err != nil {
				log.Fatal(err)
			}
			recvChannel <- message.Data
		} else if (messageType == "ConnectionDrainingMessage") {
			drainChannel <- ConnectionDrainingMessage{}
		}
		//TODO: Handle InboundACKMessage and NACKMessage
	}
}

//StartRecv starts listening for Firebase Cloud Messaging messages in a goroutine of client.recv.
func (client *FirebaseClient) StartRecv(drainChannel chan<- ConnectionDrainingMessage) <-chan SMSMessage {
	recvChannel := make(chan SMSMessage)
	go client.recv(recvChannel, drainChannel)
	return recvChannel
}

//Send sends a message to FirebaseXMPP
func (client *FirebaseClient) Send(chat xmpp.Chat) (int, error) {
	return client.xmppClient.Send(chat)
}

//ConstructAndSend constructs a xmpp.Chat object and send it using Send
func (client *FirebaseClient) ConstructAndSend(messageType, text string) (int, error) {
	chat := xmpp.Chat {
		Remote: fcmServer,
		Type: messageType,
		Text: text,
		Stamp: time.Now(),
	}
	return client.Send(chat)
}

func (client *FirebaseClient) sendACK(message UpstreamMessage) (int, error) {
	payload := ConstructACK(message.From, message.MessageID)
	return client.xmppClient.SendOrg(string(payload))
}
