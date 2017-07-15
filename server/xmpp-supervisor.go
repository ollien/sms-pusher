package main

import "./firebase-xmpp"
import "github.com/satori/go.uuid"

//StartFirebaseClient will add a client to the clients map and begin listening for connection draining messages
func StartFirebaseClient(clients map[string]firebasexmpp.FirebaseClient, configPath string) <-chan firebasexmpp.SMSMessage{
	clientID := uuid.NewV4().String()
	client := firebasexmpp.NewFirebaseClient(configPath, clientID)
	clients[clientID] = client
	drainChannel := make(chan firebasexmpp.ConnectionDrainingMessage)
	closeChannel := make(chan *firebasexmpp.FirebaseClient)
	messageChannel := client.StartRecv(drainChannel, closeChannel)
	go handleConnectionDraining(drainChannel, clients, clientID, configPath)
	go handleConnectionClose(closeChannel, clients)
	return messageChannel
}

//StartFirebaseClientOnExistingMessageChannel is identical to StartFirebaseClient but it takes a messageChannel as an argument, and will direct all messages to that channel.
func StartFirebaseClientOnExistingMessageChannel(clients map[string] firebasexmpp, configPath string, messageChannel chan<- firebasexmpp.SMSMessage) {
	clientID := uuid.NewV4().String()
	client := firebasexmpp.NewFirebaseClient(configPath, clientID)
	clients[clientID] = client
	drainChannel := make(chan firebasexmpp.ConnectionDrainingMessage)
	closeChannel := make(chan *firebasexmpp.FirebaseClient)
	client.StartRecvOnExistingChannel(drainChannel, closeChannel, messageChannel)
	go handleConnectionDraining(drainChannel, clients, clientID, configPath)
	go handleConnectionClose(closeChannel, clients)
	return messageChannel
}

func handleConnectionDraining(drainChannel <-chan firebasexmpp.ConnectionDrainingMessage, clients map[string]firebasexmpp.FirebaseClient, clientID string, configPath string) {
	_ = <- drainChannel
	StartFirebaseClient(clients, configPath)
}

func handleConnectionClose(closeChannel <-chan *firebasexmpp.FirebaseClient, clients map[string]firebasexmpp.FirebaseClient) {
	closingClient := <- closeChannel
	delete(clients, closingClient.ClientID)
}
