package firebasexmpp

import (
	"encoding/json"
	"errors"
	"fmt"
)

//FCMMessage represents a single message from FCM
type FCMMessage interface {
	PerformAction(client *FirebaseClient) error
}

//UpstreamMessage stores the basic data from any upstream Firebase Cloud Messaging XML Message.
//This isn't as general as it could be. Because the app only sends SMS/MMSes upstream, TextMessage is included in UpstreamMessage.
type UpstreamMessage struct {
	From      string `json:"from"`
	TTL       int    `json:"time_to_live"`
	MessageID string `json:"message_id"`
	Category  string `json:"category"`
	Data      json.RawMessage
}

//InboundACKMessage stores the basic data from an ACK message that Firebase CLoud Messaging when we send a message downstream.
type InboundACKMessage struct {
	From      string `json:"from"`
	MessageID string `json:"message_id"`
}

//NACKMessage stores the basic data from an NACK message that Firebase CLoud Messaging when we send a message downstream.
type NACKMessage struct {
	From             string `json:"from"`
	MessageID        string `json:"message_id"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

//ConnectionDrainingMessage indicates a CONNECTION_DRAINING message.
type ConnectionDrainingMessage struct{}

//parseFCMMessage determines the type of message that Firebase Cloud Messaging has sent upstream.
func parseFCMMessage(data []byte) (FCMMessage, error) {
	message := struct {
		MessageType string `json:"message_type"`
		Other       json.RawMessage
	}{}
	err := json.Unmarshal(data, &message)
	if err != nil {
		return nil, err
	}

	var parsedMessage FCMMessage
	switch message.MessageType {
	case "":
		//Upstream Messages have no message type. thus, if MessageType is nil, the message therefore has no message type, and we can assume it's an upstream mesage.
		parsedMessage = &UpstreamMessage{}
	case "ack":
		parsedMessage = &InboundACKMessage{}
	case "nack":
		parsedMessage = &NACKMessage{}
	case "control":
		//Per the spec, CONNECTION_DRAINING is the only control_type supported. We can save CPU time by not checking the control_type.
		parsedMessage = &ConnectionDrainingMessage{}
	default:
		return nil, errors.New("Unknown message type")
	}
	err = json.Unmarshal(data, &parsedMessage)
	if err != nil {
		return nil, err
	}

	return parsedMessage, nil
}

//PerformAction sends the SMS message upstream to the main program
func (message UpstreamMessage) PerformAction(client *FirebaseClient) error {
	_, err := client.sendACK(message)
	if err != nil {
		return err
	}

	client.recvChannel <- message

	return nil
}

//PerformAction is a stub to satisfy the FCMMessage interface.
//Presently, there is no mechanism in place to wait for an ACK, so we're just stubbing this.
func (message InboundACKMessage) PerformAction(client *FirebaseClient) error {
	return nil
}

//PerformAction will return a properly formatted error object for the error given by FCM.
func (message NACKMessage) PerformAction(client *FirebaseClient) error {
	return fmt.Errorf("firebasexmpp: %s - %s", message.Error, message.ErrorDescription)
}

//PerformAction informs the signal channel that this connection needs to be drained.
func (message ConnectionDrainingMessage) PerformAction(client *FirebaseClient) error {
	drainSignal := NewConnectionDrainingSignal(client)
	client.signalChannel <- drainSignal
	return nil
}
