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
	drainChannel  chan firebasexmpp.ConnectionDrainingSignal
}

//NewXMPPSupervisor creates a new XMPPSupervisor and starts the necessary handlers.
func NewXMPPSupervisor(configPath string) XMPPSupervisor {
	supervsior := XMPPSupervisor{
		clients:       make(map[string]firebasexmpp.FirebaseClient),
		ConfigPath:    configPath,
		signalChannel: make(chan firebasexmpp.Signal),
		spawnChannel:  make(chan chan firebasexmpp.SMSMessage),
		closeChannel:  make(chan firebasexmpp.ConnectionClosedSignal),
		drainChannel:  make(chan firebasexmpp.ConnectionDrainingSignal),
	}

	//Launch handlers
	go supervsior.listenAndSpawn()
	go supervsior.listenForSignal()
	go supervsior.listenForDraining()
	go supervsior.listenForClose()
}

//SpawnClient spawns a new FirebaseClient
func (supervisor *XMPPSupervisor) SpawnClient(messageChannel chan firebasexmpp.SMSMessage) {
	clientID := uuid.NewV4().String()
	firebaseClient := firebasexmpp.NewFirebaseClient(supervisor.configPath, signalChannel, clientID, client.spawnChannel)
	supervisor.clients[clientID] = firebaseClient
}

//listenAndSpawns listens on supervisor.spawnChannel and spawns clients as necessary
//Exits when supervisor.spawnChannel is closed
func (supervisor *XMPPSupervisor) listenAndSpawn() {
	for messageChannel := range spawnChannel {
		supervisor.SpawnClient(messageChannel)
	}
}

//listenForSignal listens on supervisor.signalChannel and passes the signal aloong to the appropriate channels
//Exists when supervisor.spawnChannel closes
func (supervisor *XMPPSupervisor) listenForSignal() {
	for signal := range signalChannel {
		switch signal.(type) {
		case firebasexmpp.ConnectionDrainingSignal:
			closeChannel <- signal
		case firebasexmpp.ConnectionClosedSignal:
			drainChannel <- signal
		}
	}
}

//listenForDraining listens on supervisor.drainChannel and spawns a new client as necessary
//Exists when supervisor.drainChannel is closed
func (supervisor *XMPPSupervisor) listenForDraining() {
	for signal := range drainChannel {
		supervisor.spawnChannel <- signal.messageChannel
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
