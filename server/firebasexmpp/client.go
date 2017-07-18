package firebasexmpp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-xmpp"
)

const fcmServer = "fcm-xmpp.googleapis.com"
const fcmDevPort = 5236
const fcmProdPort = 5235
const fcmUsernameAddres = "gcm.googleapis.com"

//FirebaseClient stores the data necessary to be an XMPP Client for Firebase Cloud Messaging. See the spec at https://firebase.google.com/docs/cloud-messaging/xmpp-server-ref
//senderID and severKey refer to their corresponding FCM properties. ClientID is simply an id to identify clients. It can safely be ommitted, but your connectionClosedCallback will receive an empty string
type FirebaseClient struct {
	xmppClient    xmpp.Client
	ClientID      string
	senderID      string
	serverKey     string
	signalChannel chan<- Signal
}

//Config stores the details necessary for authenticating to Firebase Cloud Messaging's XMPP server, which cannot be hardcoded or put into version control.
type Config struct {
	SenderID  string
	ServerKey string
}

//NewFirebaseClient creates a FirebaseClient from configuration file.
func NewFirebaseClient(configPath string, clientID string, signalChannel chan<- Signal) FirebaseClient {
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
		xmppClient:    *client,
		ClientID:      clientID,
		senderID:      config.SenderID,
		serverKey:     config.SererKey,
		signalChannel: signalChannel,
	}
}

//recv listens for incomgin messages from Firebase Cloud Messaging.
func (client *FirebaseClient) recv(recvChannel chan<- SMSMessage) {
	for {
		data, err := client.xmppClient.Recv()
		if err != nil {
			//encoding/xml adds a bunch of extra stuff to XML errors, including the line number. However, all we care about is whether or not an EOF was reached.
			if strings.Contains(err.Error(), io.EOF.Error()) {
				closeSignal := NewConnectionClosedSignal(client)
				client.signalChannel <- &closeSignal
				break
			} else {
				log.Fatal(err)
			}
		}
		chat, ok := data.(xmpp.Chat)
		if !ok {
			//xmpp.Recv can return a xmpp.Chat or a xmpp.Presence. We don't care about presence notifications.
			//Though the FCM spec makes no mention of them, because it doesn't explicitly say we will never recieve them, we must handle them somehow - in this case, ignoring them.
			continue
		}
		messageBody := []byte(chat.Other[0])
		messageType, err := GetMessageType(messageBody)
		if err != nil {
			//Don't need to quit for unknowm message types
			log.Println(err)
		} else if messageType == "UpstreamMessage" {
			var message UpstreamMessage
			json.Unmarshal(messageBody, &message)
			_, err := client.sendACK(message)
			if err != nil {
				log.Fatal(err)
			}
			recvChannel <- message.Data
		} else if messageType == "ConnectionDrainingMessage" {
			drainSignal := NewConnectionDrainingSignal(client)
			client.signalChannel <- &drainSignal
		}
		//TODO: Handle InboundACKMessage and NACKMessage
	}
}

//StartRecv starts listening for Firebase Cloud Messaging messages in a goroutine of client.recv.
func (client *FirebaseClient) StartRecv() chan SMSMessage {
	recvChannel := make(chan SMSMessage)
	client.StartRecvOnExistingChannel(recvChannel)
	return recvChannel
}

//StartRecvOnExistingChannel is identical to StartRecv, except that it takes a recvChannel as an argument, and will direct all messages to that channel.
func (client *FirebaseClient) StartRecvOnExistingChannel(recvChannel chan SMSMessage) {
	go client.recv(recvChannel)
}

//Send sends a message to FirebaseXMPP
func (client *FirebaseClient) Send(chat xmpp.Chat) (int, error) {
	return client.xmppClient.Send(chat)
}

//ConstructAndSend constructs a xmpp.Chat object and send it using Send
func (client *FirebaseClient) ConstructAndSend(messageType, text string) (int, error) {
	chat := xmpp.Chat{
		Remote: fcmServer,
		Type:   messageType,
		Text:   text,
		Stamp:  time.Now(),
	}
	return client.Send(chat)
}

func (client *FirebaseClient) sendACK(message UpstreamMessage) (int, error) {
	payload := ConstructACK(message.From, message.MessageID)
	return client.xmppClient.SendOrg(string(payload))
}
