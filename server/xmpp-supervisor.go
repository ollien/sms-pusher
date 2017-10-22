package main

import (
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/satori/go.uuid"
)

//XMPPSupervisor supervises all Firebase XMPP connections
type XMPPSupervisor struct {
	clients       map[string]ClientContainer
	ConfigPath    string
	signalChannel chan firebasexmpp.Signal
	spawnChannel  chan ClientContainer
	closeChannel  chan *firebasexmpp.ConnectionClosedSignal
	drainChannel  chan *firebasexmpp.ConnectionDrainingSignal
}

//ClientContainer holds a client and its channels
type ClientContainer struct {
	client       firebasexmpp.FirebaseClient
	sendChannel  chan interface{}
	errorChannel chan error
	recvChannel  chan firebasexmpp.SMSMessage
}

//NewXMPPSupervisor creates a new XMPPSupervisor and starts the necessary handlers.
func NewXMPPSupervisor(configPath string) XMPPSupervisor {
	supervisor := XMPPSupervisor{
		clients:       make(map[string]ClientContainer),
		ConfigPath:    configPath,
		signalChannel: make(chan firebasexmpp.Signal),
		spawnChannel:  make(chan ClientContainer),
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
func (supervisor *XMPPSupervisor) SpawnClient(messageChannel chan firebasexmpp.SMSMessage, sendChannel chan interface{}, errorChannel chan error) {
	container := ClientContainer{
		sendChannel:  sendChannel,
		errorChannel: errorChannel,
		recvChannel:  messageChannel,
	}
	supervisor.spawnClientFromContainer(container)
}

func (supervisor *XMPPSupervisor) spawnClientFromContainer(container ClientContainer) {
	clientID := uuid.NewV4().String()
	firebaseClient := firebasexmpp.NewFirebaseClient(supervisor.ConfigPath, clientID, supervisor.signalChannel)
	container.client = firebaseClient
	supervisor.clients[container.client.ClientID] = container
	go container.client.StartRecv(container.recvChannel)
	go container.client.ListenForSend(container.sendChannel, container.errorChannel)
}

//listenAndSpawns listens on supervisor.spawnChannel and spawns clients as necessary
//Exits when supervisor.spawnChannel is closed
func (supervisor *XMPPSupervisor) listenAndSpawn() {
	for container := range supervisor.spawnChannel {
		supervisor.spawnClientFromContainer(container)
	}
}

//listenForSignal listens on supervisor.signalChannel and passes the signal aloong to the appropriate channels
//Exists when supervisor.spawnChannel closes
func (supervisor *XMPPSupervisor) listenForSignal() {
	for signal := range supervisor.signalChannel {
		switch convertedSignal := signal.(type) {
		case *firebasexmpp.ConnectionDrainingSignal:
			supervisor.drainChannel <- convertedSignal
		case *firebasexmpp.ConnectionClosedSignal:
			supervisor.closeChannel <- convertedSignal
		}
	}
}

//listenForDraining listens on supervisor.drainChannel and spawns a new client as necessary
//Exists when supervisor.drainChannel is closed
func (supervisor *XMPPSupervisor) listenForDraining() {
	for signal := range supervisor.drainChannel {
		supervisor.spawnChannel <- supervisor.clients[signal.Client.ClientID]
	}
}

//ListenForeClose listens on supervisor.closeChannel and deletes closed clients from supervisor.clients as necessary
//Exits when supervisor.closeChannel is closed
func (supervisor *XMPPSupervisor) listenForClose() {
	for signal := range supervisor.closeChannel {
		clientID := signal.Client.ClientID
		delete(supervisor.clients, clientID)
	}
}
