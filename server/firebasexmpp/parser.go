package firebasexmpp

import (
	"encoding/json"
	"errors"
)

//TextMessage represents either a SMS or an MMS.
type TextMessage interface {
	isMMS() bool
}

//SMSMessage stores the data sent by the app upstream about incoming SMS messages.
type SMSMessage struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message,omitempty"`
	Timestamp   int64  `json:"timestamp,string"`
}

//MMSMessage represents an MMS message that comes in
type MMSMessage struct {
	SMSMessage
	Recipients []string `json:"recipients"`
	//Holds any message components that aren't just a message.
	Parts []MMSPart `json:"parts,string"`
}

//MMSPart represents a part of an MMS message.
type MMSPart struct {
	PartType string `json:"type"`
	Data     []byte `json:"data"`
}

//UnknownMessage represents a message a message of undetermined type
type UnknownMessage struct {
	MessageType *string `json:"message_type"`
	Other       json.RawMessage
}

//UpstreamMessage stores the basic data from any upstream Firebase Cloud Messaging XML Message.
//This isn't as general as it could be. Because the app only sends SMS/MMSes upstream, TextMessage is included in UpstreamMessage.
type UpstreamMessage struct {
	From      string `json:"from"`
	TTL       int    `json:"time_to_live"`
	MessageID string `json:"message_id"`
	Category  string `json:"category"`
	Data      TextMessage
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

//GetMessageType determines the type of message that Firebase Cloud Messaging has sent upstream.
func GetMessageType(rawData []byte) (string, error) {
	message := UnknownMessage{}
	err := json.Unmarshal(rawData, &message)
	if err != nil {
		return "", err
	}

	//Upstream Messages have no message type. thus, if MessageType is nil, the message therefore has no message type, and we can assume it's an upstream mesage.
	if message.MessageType == nil {
		return "UpstreamMessage", nil
	}

	switch *message.MessageType {
	case "ack":
		return "InboundACKMessage", nil
	case "nack":
		return "NACKMessage", nil
	case "control":
		//Per the spec, CONNECTION_DRAINING is the only control_type supported. We can save CPU time by not checking the control_type.
		return "ConnectionDrainingMessage", nil
	default:
		return *message.MessageType, errors.New("Unknown message type")
	}
}

func (message *SMSMessage) isMMS() {
	return false
}

func (message *MMSMessage) isMMS() {
	return len(messge.recipients) > 1 && len(message.Parts) != nil
}
