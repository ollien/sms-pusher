package firebasexmpp

//SignalType represents the type of signal produced by FCM
type SignalType int

const (
	//ConnectionDrainingSignal represents a CONNCETION_DRAINING signal from FCM, which tells us to prepare for this connection to be closed.
	ConnectionDrainingSignal SignalType = iota
	//ConnectionClosedSignal represents a signal to the supervisor that this connection is closed.
	ConnectionClosedSignal
)

//Signal represents a signal to the XMPP supervisor
type Signal interface {
	init(client *FirebaseClient)
}

//ConnectionDrainingSignal is a signal to represeent a CONNECTION_DRAINING message
//Implements Signal interface
type ConnectionDrainingSignal struct {
	Client *FirebaseClient
}

//ConnectionClosedSignal is a signal to represent the socket being closed
//Implements Signal interface
type ConnectionClosedSignal struct {
	Client *FirebaseClient
}

//NewConnectionDrainingSignal generates a new ConnectionDrainingSignal
func NewConnectionDrainingSignal(client *FirebaseClient) ConnectionDrainingSignal {
	signal := ConnectionDrainingSignal{}
	signal.init(client)
	return signal
}

func (signal *ConnectionDrainingSignal) init(client *FirebaseClient) {
	signal.Client = client
}

//NewConnectionClosedSignal generates a new ConnectionClosedSignal
func NewConnectionClosedSignal(client *FirebaseClient) ConnectionClosedSignal {
	signal := ConnectionClosedSignal{}
	signal.init(client)
	return signal
}

func (signal *ConnectionClosedSignal) init(client *FirebaseClient) {
	signal.Client = client
}
