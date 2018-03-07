package main

import (
	"github.com/ollien/sms-pusher/server/config"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/satori/go.uuid"
)

//XMPPSupervisor supervises all Firebase XMPP connections
type XMPPSupervisor struct {
	clients       map[string]ClientContainer
	Config        config.XMPPConfig
	SendChannel   chan firebasexmpp.OutboundMessage
	signalChannel chan firebasexmpp.Signal
	spawnChannel  chan ClientContainer
}

//ClientContainer holds a client and its channels
type ClientContainer struct {
	client       firebasexmpp.FirebaseClient
	sendChannel  chan firebasexmpp.OutboundMessage
	errorChannel chan firebasexmpp.ClientError
	recvChannel  chan firebasexmpp.SMSMessage
}

//NewXMPPSupervisor creates a new XMPPSupervisor and starts the necessary handlers.
func NewXMPPSupervisor(xmppConfig config.XMPPConfig) XMPPSupervisor {
	supervisor := XMPPSupervisor{
		clients:       make(map[string]ClientContainer),
		Config:        xmppConfig,
		signalChannel: make(chan firebasexmpp.Signal),
		spawnChannel:  make(chan ClientContainer),
	}

	//Launch handlers
	go supervisor.listenAndSpawn()
	go supervisor.listenForSignal()

	return supervisor
}

//SpawnClient spawns a new FirebaseClient
func (supervisor *XMPPSupervisor) SpawnClient(messageChannel chan firebasexmpp.SMSMessage, sendChannel chan firebasexmpp.OutboundMessage) {
	container := ClientContainer{
		sendChannel:  sendChannel,
		errorChannel: make(chan xmpp.ClientError),
		recvChannel:  messageChannel,
	}
	supervisor.spawnClientFromContainer(container)
}

func (supervisor *XMPPSupervisor) spawnClientFromContainer(container ClientContainer) {
	clientID := uuid.NewV4().String()
	firebaseClient := firebasexmpp.NewFirebaseClient(supervisor.Config, clientID, supervisor.signalChannel, container.errorChannel)
	container.client = firebaseClient
	supervisor.clients[container.client.ClientID] = container
	go container.client.StartRecv(container.recvChannel)
	go container.client.ListenForSend(container.sendChannel)
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
