package firebasexmpp

//Signal represents a signal to the XMPP supervisor
type Signal interface {
	init(client *FirebaseClient)
}

//ConnectionDrainingSignal is a signal to represeent a CONNECTION_DRAINING message
//Implements Signal interface
type ConnectionDrainingSignal struct {
	client         *FirebaseClient
	MessageChannel chan SMSMessage
}

//ConnectionClosedSignal is a signal to represent the socket being closed
//Implements Signal interface
type ConnectionClosedSignal struct {
	client *FirebaseClient
}

//NewConnectionDrainingSignal generates a new ConnectionDrainingSignal
func NewConnectionDrainingSignal(client *FirebaseClient, messageChannel chan SMSMessage) ConnectionDrainingSignal {
	signal := ConnectionDrainingSignal{}
	signal.init(client)
	signal.MessageChannel = messageChannel
	return signal
}

func (signal *ConnectionDrainingSignal) init(client *FirebaseClient) {
	signal.client = client
}

//NewConnectionClosedSignal generates a new ConnectionClosedSignal
func NewConnectionClosedSignal(client *FirebaseClient) ConnectionClosedSignal {
	signal := ConnectionClosedSignal{}
	signal.init(client)
	return signal
}

func (signal *ConnectionClosedSignal) init(client *FirebaseClient) {
	signal.client = client
}
