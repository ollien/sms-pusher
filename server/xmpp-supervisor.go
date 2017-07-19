package main

import (
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/satori/go.uuid"
)

//XMPPSupervisor supervises all Firebase XMPP connections
type XMPPSupervisor struct {
	clients       map[string]firebasexmpp.FirebaseClient
	ConfigPath    string
	signalChannel chan firebasexmpp.Signal
	spawnChannel  chan chan firebasexmpp.SMSMessage
	closeChannel  chan *firebasexmpp.ConnectionClosedSignal
	drainChannel  chan *firebasexmpp.ConnectionDrainingSignal
}

//NewXMPPSupervisor creates a new XMPPSupervisor and starts the necessary handlers.
func NewXMPPSupervisor(configPath string) XMPPSupervisor {
	supervisor := XMPPSupervisor{
		clients:       make(map[string]firebasexmpp.FirebaseClient),
		ConfigPath:    configPath,
		signalChannel: make(chan firebasexmpp.Signal),
		spawnChannel:  make(chan chan firebasexmpp.SMSMessage),
		closeChannel:  make(chan *firebasexmpp.ConnectionClosedSignal),
		drainChannel:  make(chan *firebasexmpp.ConnectionDrainingSignal),
	}

	//Launch handlers
	go supervisor.listenAndSpawn()
	go supervisor.listenForSignal()
	go supervisor.listenForDraining()
	go supervisor.listenForClose()

	return supervisor
}

//SpawnClient spawns a new FirebaseClient
func (supervisor *XMPPSupervisor) SpawnClient(messageChannel chan firebasexmpp.SMSMessage) {
	clientID := uuid.NewV4().String()
	firebaseClient := firebasexmpp.NewFirebaseClient(supervisor.ConfigPath, clientID, supervisor.signalChannel)
	supervisor.clients[clientID] = firebaseClient
}

//listenAndSpawns listens on supervisor.spawnChannel and spawns clients as necessary
//Exits when supervisor.spawnChannel is closed
func (supervisor *XMPPSupervisor) listenAndSpawn() {
	for messageChannel := range supervisor.spawnChannel {
		supervisor.SpawnClient(messageChannel)
	}
}

//listenForSignal listens on supervisor.signalChannel and passes the signal aloong to the appropriate channels
//Exists when supervisor.spawnChannel closes
func (supervisor *XMPPSupervisor) listenForSignal() {
	for signal := range supervisor.signalChannel {
		switch convertedSignal := signal.(type) {
		case *firebasexmpp.ConnectionDrainingSignal:
			supervisor.closeChannel <- convertedSignal
		case *firebasexmpp.ConnectionClosedSignal:
			supervisor.drainChannel <- convertedSignal
		}
	}
}

//listenForDraining listens on supervisor.drainChannel and spawns a new client as necessary
//Exists when supervisor.drainChannel is closed
func (supervisor *XMPPSupervisor) listenForDraining() {
	for signal := range supervisor.drainChannel {
		supervisor.spawnChannel <- signal.MessageChannel
	}
}

//ListenForeClose listens on supervisor.closeChannel and deletes closed clients from supervisor.clients as necessary
//Exits when supervisor.closeChannel is closed
func (supervisor *XMPPSupervisor) listenForClose() {
	for signal := range supervisor.closeChannel {
		clientID := signal.client.clientID
		delete(supervisor.clients, clientID)
	}
}
