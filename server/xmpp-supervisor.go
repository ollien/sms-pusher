package main

import "./firebase-xmpp"
import "github.com/satori/go.uuid"

//StartFirebaseClient will add a client to the clients map and begin listening for connection draining messages
func StartFirebaseClient(clients map[string]firebasexmpp.FirebaseClient, configPath string) {
	client := firebasexmpp.NewFirebaseClient(configPath)
	clientID := uuid.NewV4().String()
	clients[clientID] = client
	drainChannel := make(chan firebasexmpp.ConnectionDrainingMessage)
	messageChannel := client.StartRecv(drainChannel)
}

func handleConnectionDraining(drainChannel <-chan firebasexmpp.ConnectionDrainingMessage, clients map[string]firebasexmpp.FirebaseClient, clientID string, configPath string) {
	_ = <- drainChannel
	StartFirebaseClient(clients, configPath)
}
