package main

import (
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const clientErrorFormat = "Client %s: %s"

//XMPPSupervisor supervises all Firebase XMPP connections
type XMPPSupervisor struct {
	clients       map[string]ClientContainer
	logger        *logrus.Logger
	recvChannel   chan firebasexmpp.UpstreamMessage
	sendChannel   chan firebasexmpp.DownstreamPayload
	signalChannel chan firebasexmpp.Signal
	spawnChannel  chan ClientContainer
}

//ClientContainer holds a client and its channels
type ClientContainer struct {
	client       firebasexmpp.FirebaseClient
	logger       *logrus.Logger
	errorChannel chan firebasexmpp.ClientError
}

//NewXMPPSupervisor creates a new XMPPSupervisor and starts the necessary handlers, given the channels to receive messages from firebase, and the channels to send messages to firebase.
func NewXMPPSupervisor(recvChannel chan firebasexmpp.UpstreamMessage, sendChannel chan firebasexmpp.DownstreamPayload, logger *logrus.Logger) XMPPSupervisor {
	supervisor := XMPPSupervisor{
		clients:       make(map[string]ClientContainer),
		logger:        logger,
		signalChannel: make(chan firebasexmpp.Signal),
		recvChannel:   recvChannel,
		sendChannel:   sendChannel,
		spawnChannel:  make(chan ClientContainer),
	}

	//Launch handlers
	go supervisor.listenAndSpawn()
	go supervisor.listenForSignal()

	return supervisor
}

//SpawnClient spawns a new FirebaseClient
func (supervisor *XMPPSupervisor) SpawnClient() error {
	container := ClientContainer{
		logger:       supervisor.logger,
		errorChannel: make(chan firebasexmpp.ClientError),
	}

	return supervisor.spawnClientFromContainer(container)
}

func (supervisor *XMPPSupervisor) spawnClientFromContainer(container ClientContainer) error {
	//Because errors can happen during creation, we need to start the listening for errors routine now.
	rawClientID, err := uuid.NewV4()
	if err != nil {
		return err
	}
	clientID := rawClientID.String()

	firebaseClient, err := firebasexmpp.NewFirebaseClient(clientID, supervisor.recvChannel, supervisor.sendChannel, supervisor.signalChannel, container.errorChannel)
	if err != nil {
		return err
	}

	container.client = firebaseClient
	supervisor.clients[container.client.ClientID] = container
	go container.listenForError()
	go container.client.StartRecv()
	go container.client.ListenForSend()

	return nil
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

//listenForError listens on a client's error channel and logs it to the logrus logger.
//Exits when container.errorChannel closes
func (container ClientContainer) listenForError() {
	for clientError := range container.errorChannel {
		if clientError.Fatal {
			container.logger.Fatalf(clientErrorFormat, container.client.ClientID, clientError.Err)
		} else {
			container.logger.Errorf(clientErrorFormat, container.client.ClientID, clientError.Err)
		}
	}
}
