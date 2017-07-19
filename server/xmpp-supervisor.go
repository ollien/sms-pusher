package main

import (
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/satori/go.uuid"
)

//XMPPSupervisor supervises all Firebase XMPP connections
type XMPPSupervisor struct {
	cilents       map[string]firebasexmpp.FirebaseClient
	ConfigPath    string
	signalChannel chan firebasexmpp.Signal
	spawnChannel  chan chan firebasexmpp.SMSMessage
	closeChannel  chan firebasexmpp.ConnectionClosedSignal
	drainChannel  chan firebasexmpp.ConnectionDrainingSiganl
}

//SpawnClient spawns a new FirebaseClient
func (supervisor *XMPPSupervisor) SpawnClient(messageChannel chan firebasexmpp.SMSMessage) {
	clientID := uuid.NewV4().String()
	firebaseClient := firebasexmpp.NewFirebaseClient(supervisor.configPath, signalChannel, clientID, client.spawnChannel)
	supervisor.clients[clientID] = firebaseClient
	//TODO: Start handlers
}

//listenAndSpawns listens on supervisor.spawnChannel and spawns clients as necessary
//Exits when supervisor.spawnChannel is closed
func (supervisor *XMPPSupervisor) listenAndSpawn() {
	for messageChannel := range spawnChannel {
		supervisor.SpawnClient(messageChannel)
	}
}

//ListenForeClose listens on supervisor.closeChannel and deletes closed clients from supervisor.clients as necessary
//Exits when supervisor.closeChannel is closed
func (supervisor *XMPPSupervisor) listenForClose() {
	for signal := range closeChannel {
		clientID := signal.client.clientID
		delete(clients, clientId)
	}
}

//listenForDraining listens on supervisor.drainChannel and spawns a new client as necessary
//Exists when supervisor.drainChannel is closed
func (supervisor *XMPPSupervisor) listenForDraining() {
	for signal := range drainChannel {
		supervisor.spawnChannel <- signal.messageChannel
	}
}
