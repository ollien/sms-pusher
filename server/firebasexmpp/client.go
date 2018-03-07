package firebasexmpp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mattn/go-xmpp"
	"github.com/ollien/sms-pusher/server/config"
)

const fcmServer = "fcm-xmpp.googleapis.com"
const fcmDevPort = 5236
const fcmProdPort = 5235
const fcmUsernameAddres = "gcm.googleapis.com"

//FirebaseClient stores the data necessary to be an XMPP Client for Firebase Cloud Messaging. See the spec at https://firebase.google.com/docs/cloud-messaging/xmpp-server-ref
//senderID and severKey refer to their corresponding FCM properties. ClientID is simply an id to identify clients. It can safely be ommitted, but your connectionClosedCallback will receive an empty string
//Note that Signal will recieve pointer types of signals such as *ConnectionDrainingSignal and *ConnectionClosedSignal rather than ConnectionDrainingSignal and ConnectionClosedSignal respectively.
type FirebaseClient struct {
	xmppClient    xmpp.Client
	ClientID      string
	senderID      string
	serverKey     string
	signalChannel chan<- Signal
	errorChannel  chan<- ClientError
}

//ClientError represents an error that occurs within a cient
type ClientError struct {
	Err   error
	Fatal bool
}

//Config stores the details necessary for authenticating to Firebase Cloud Messaging's XMPP server, which cannot be hardcoded or put into version control.
type Config struct {
	SenderID  string
	ServerKey string
}

//NewFirebaseClient creates a FirebaseClient from the given XMPPConfig
func NewFirebaseClient(xmppConfig config.XMPPConfig, clientID string, signalChannel chan<- Signal, errorChannel chan<- ClientError) FirebaseClient {
	//TODO: Detect if debug. For now, use dev port and set debug to true. For now, we will just always do this.
	server := fmt.Sprintf("%s:%d", fcmServer, fcmDevPort)
	username := fmt.Sprintf("%s@%s", xmppConfig.SenderID, fcmUsernameAddres)
	client, err := xmpp.NewClient(server, username, xmppConfig.ServerKey, true)
	if err != nil {
		log.Fatal(err)
	}

	return FirebaseClient{
		xmppClient:    *client,
		ClientID:      clientID,
		senderID:      xmppConfig.SenderID,
		serverKey:     xmppConfig.ServerKey,
		signalChannel: signalChannel,
		errorChannel:  errorChannel,
	}
}

//StartRecv listens for incoming  messages from Firebase Cloud Messaging and sends them to recvChannel.
func (client *FirebaseClient) StartRecv(recvChannel chan SMSMessage) {
	for {
		data, err := client.xmppClient.Recv()
		if err != nil {
			//encoding/xml adds a bunch of extra stuff to XML errors, including the line number. However, all we care about is whether or not an EOF was reached.
			if strings.Contains(err.Error(), io.EOF.Error()) {
				closeSignal := NewConnectionClosedSignal(client)
				client.signalChannel <- closeSignal
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
			client.signalChannel <- drainSignal
		}
		//TODO: Handle InboundACKMessage and NACKMessage
	}
}

//ListenForSend listens for a message on sendChannel and sends the message.
//Terminates when sendChannel is closed
func (client *FirebaseClient) ListenForSend(sendChannel <-chan OutboundMessage) {
	for message := range sendChannel {
		_, err := message.Send(client.xmppClient)
		if err != nil {
			errorToSend := ClientError{
				Err:   err,
				Fatal: false,
			}
			client.errorChannel <- errorToSend
		}
	}
}

func (client *FirebaseClient) sendACK(message UpstreamMessage) (int, error) {
	ack, err := ConstructACK(message.From, message.MessageID)
	if err != nil {
		return 0, err
	}

	return ack.Send(client.xmppClient)
}
