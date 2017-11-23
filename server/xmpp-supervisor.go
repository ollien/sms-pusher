package main

import (
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/satori/go.uuid"
)

//XMPPSupervisor supervises all Firebase XMPP connections
type XMPPSupervisor struct {
	clients       map[string]ClientContainer
	ConfigPath    string
	SendChannel   chan firebasexmpp.OutboundMessage
	signalChannel chan firebasexmpp.Signal
	spawnChannel  chan ClientContainer
}

//ClientContainer holds a client and its channels
type ClientContainer struct {
	client       firebasexmpp.FirebaseClient
	sendChannel  chan firebasexmpp.OutboundMessage
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
	}

	//Launch handlers
	go supervisor.listenAndSpawn()
	go supervisor.listenForSignal()

	return supervisor
}

//SpawnClient spawns a new FirebaseClient
func (supervisor *XMPPSupervisor) SpawnClient(messageChannel chan firebasexmpp.SMSMessage, sendChannel chan firebasexmpp.OutboundMessage, errorChannel chan error) {
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

func (supervisor *XMPPSupervisor) listenForSend() {
	for message := range supervisor.SendChannel {
		for _, client := range supervisor.clients {
			client.sendChannel <- message
		}
	}
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
		if signal.Type == firebasexmpp.ConnectionDrainingSignal {
			supervisor.spawnChannel <- supervisor.clients[signal.Client.ClientID]
		} else {
			clientID := signal.Client.ClientID
			delete(supervisor.clients, clientID)
		}
	}
}
