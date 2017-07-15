package main

import "./firebase-xmpp"
import "github.com/satori/go.uuid"

//StartFirebaseClient will add a client to the clients map and begin listening for connection draining messages
func StartFirebaseClient(clients map[string]firebasexmpp.FirebaseClient, configPath string) <-chan firebasexmpp.SMSMessage{
	clientID := uuid.NewV4().String()
	clients[clientID] = client
	client := firebasexmpp.NewFirebaseClient(configPath, clientID)
	drainChannel := make(chan firebasexmpp.ConnectionDrainingMessage)
	closeChannel := make(chan *firebasexmpp.FirebaseClient)
	messageChannel := client.StartRecv(drainChannel, closeChannel)
	go handleConnection(drainChannel, clients, cliendId, configPath)
	return messageChannel
}

func handleConnectionDraining(drainChannel <-chan firebasexmpp.ConnectionDrainingMessage, clients map[string]firebasexmpp.FirebaseClient, clientID string, configPath string) {
	_ = <- drainChannel
	StartFirebaseClient(clients, configPath)
}
