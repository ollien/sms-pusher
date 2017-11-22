package firebasexmpp

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/mattn/go-xmpp"
	uuid "github.com/satori/go.uuid"
)

//OutboundMessage represents a single message to be sent out to the XMPP server
type OutboundMessage interface {
	Send(xmpp.Client) (int, error)
}

//MessageStanza stores the data from the message stanza in outgoing messages. Used for marshalling XML.
type MessageStanza struct {
	XMLName xml.Name `xml:"message"`
	To      string   `xml:"to,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	ID      string   `xml:"id,attr,omitempty"`
	Body    interface{}
}

//GCMStanza stores the data in the gcm stanza in outgoing messages. Used for marshalling XML.
type GCMStanza struct {
	XMLName xml.Name `xml:"gcm"`
	XMLNS   string   `xml:"xmlns,attr"`
	Value   string   `xml:",innerxml"` //A bit of a hack, but it works. Chardata escaped our JSON, but innerxml will not.
}

//BodyStanza stores the data in the body stanza in outgoing messages. Used for marshalling XML.
type BodyStanza struct {
	XMLName xml.Name `xml:"body"`
	Value   string   `xml:",chardata"`
}

//ACKPayload stores the data in the ACK payload. Used for marshalling JSON.
type ACKPayload struct {
	To          string `json:"to"`
	MessageID   string `json:"message_id"`
	MessageType string `json:"message_type"`
}

//DownstreamPayload stores the data to be sent downstream. Used for marshaling JSON.
//Because our client is Android specific, we ignore content_availabe and mutable_content, which do nothing for Android.
type DownstreamPayload struct {
	To                       string      `json:"to,omitempty"`
	Condition                string      `json:"condition,omitempty"`
	MessageID                string      `json:"message_id"`
	CollapseKey              string      `json:"collapse_key,omitempty"`
	Priority                 string      `json:"priority,omitempty"`
	TTL                      int         `json:"time_to_live,omitempty"`
	DeliveryReceiptRequested bool        `json:"delivery_receipt_requested,omitempty"`
	DryRun                   bool        `json:"dry_run,omitempty"`
	Data                     interface{} `json:"data,omitempty"`
	Notification             bool        `json:"notification,omitempty"`
}

//NewGCMStanza makes a new GCMStanza. the XMLNS should always be google:mobile:data.
func NewGCMStanza(payload string) GCMStanza {
	return GCMStanza{
		Value: payload,
		XMLNS: "google:mobile:data",
	}
}

//NewACKPayload makes a new ACKPayload. MessageType should always be ack.
func NewACKPayload(registrationID, messageID string) ACKPayload {
	return ACKPayload{
		To:          registrationID,
		MessageID:   messageID,
		MessageType: "ack",
	}
}

//ConstructACK constructs a full ACK message to be send to the server.
func ConstructACK(registrationID, messageID string) []byte {
	payload := NewACKPayload(registrationID, messageID)
	marshaledPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	messageStanza := MessageStanza{
		Body: NewGCMStanza(string(marshaledPayload)),
	}
	marshaledMessageStanza, err := xml.Marshal(messageStanza)
	if err != nil {
		log.Fatal(err)
	}
	return marshaledMessageStanza
}

//ConstructDownstreamSMS constructs a ConstreamPayload and returns the marshaled result
func ConstructDownstreamSMS(deviceTo []byte, message SMSMessage) []byte {
	messageID := uuid.NewV4()
	payload := DownstreamPayload{
		To:        string(deviceTo),
		MessageID: messageID.String(),
		Priority:  "high",
		TTL:       3600,
		Data:      message,
	}
	//TODO: Abstract this so we're not repeating this every time we need to send something downstream
	marshaledPayload, err := json.Marshal(payload)
	messageStanza := MessageStanza{
		Body: NewGCMStanza(string(marshaledPayload)),
	}
	marshaledMessageStanza, err := xml.Marshal(messageStanza)
	if err != nil {
		log.Fatal(err)
	}
	return marshaledMessageStanza
}
