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
type Signal struct {
	Client *FirebaseClient
	Type   SignalType
}
