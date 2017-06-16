package firebase_xmpp

import "encoding/json"
import "fmt"
import "github.com/mattn/go-xmpp"
import "log"
import "os"
import "time"


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
	//TODO: Detect if debug. For now, use dev port and set debug to true. For now, we will just always do this.
	server := fmt.Sprintf("%s:%d", FCM_SERVER, FCM_DEV_PORT)
	username := fmt.Sprintf("%s@%s", config.SenderId, FCM_USERNAME_ADDRESS)
	client, err := xmpp.NewClient(server, username, config.ServerKey, true)
	if err != nil {
		log.Fatal(err)
	}
	return FirebaseClient{
		xmppClient: *client,
		senderId: config.SenderId,
		serverKey: config.ServerKey,
	}
}

func (client *FirebaseClient) recv(recvChannel chan interface{}) {
	for {
		data, err := client.xmppClient.Recv()
		if err != nil {
			log.Fatal(err)
		}
		chat := data.(xmpp.Chat)
		messageBody := []byte(chat.Other[0])
		switch messageType := GetMessageType(messageBody); messageType {
			case "InboundACKMessage":
				var message InboundACKMessage
				json.Unmarshal(messageBody, &message)
				//TODO: Process ACK message so we don't just silently receive acknowledgement
			case "NACKMessage":
				var message NACKMessage
				json.Unmarshal(messageBody, &message)
				//TODO: Process NACK message so we don't just silently fail
			case "UpstreamMessage":
				var message UpstreamMessage
				json.Unmarshal(messageBody, &message)
				recvChannel <- message.Data
		}
	}
}

func (client *FirebaseClient) StartRecv() <-chan interface{} {
	recvChannel := make(chan interface{})
	go client.recv(recvChannel)
	return recvChannel
}

func (client *FirebaseClient) Send(chat xmpp.Chat) (int, error) {
	return client.xmppClient.Send(chat)
}

//Construct a xmpp.Chat object and send it using Send
func (client *FirebaseClient) ConstructAndSend(messageType, text string) (int, error) {
	chat := xmpp.Chat {
		Remote: FCM_SERVER,
		Type: messageType,
		Text: text,
		Stamp: time.Now(),
	}
	return client.Send(chat)
}

func (client *FirebaseClient) sendAck(message OutboundACKMessage) (int, error) {
	payload := fmt.Sprintf(
		`<message id="">
			<gcm xmlns="google:mobile:data"
				{
					"to": "%s",
					"message_id": "%s",
					"message_type": "ack"
				}
			</gcm>
		</message>`, message.To, message.MessageId)
	return client.ConstructAndSend("normal", payload)
}
