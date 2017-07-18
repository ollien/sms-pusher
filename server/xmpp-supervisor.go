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
	closeChannel  chan *firebasexmpp.FirebaseClient
	drainChannel  chan *firebaexmpp.FirebaseClient
}

//SpawnClient spawns a new FirebaseClient
func (supervisor *XMPPSupervisor) SpawnClient(messageChannel chan firebasexmpp.SMSMessage) {
	clientID := uuid.NewV4().String()
	firebaseClient := firebasexmpp.NewFirebaseClient(supervisor.configPath, signalChannel, clientID, client.spawnChannel)
	supervisor.clients[clientID] = firebaseClient
	//TODO: Start handlers
}
