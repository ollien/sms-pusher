package firebasexmpp

import (
	"encoding/json"
	"encoding/xml"
	"log"
)

//MessageStanza stores the data from the message stanza in outgoing messages. Used for marshalling XML.
type MessageStanza struct {
	XMLName xml.Name `xml:"message"`
	To string `xml:"to,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	ID string `xml:"id,attr,omitempty"`
	Body interface{}
}

//GCMStanza stores the data in the gcm stanza in outgoing messages. Used for marshalling XML.
type GCMStanza struct {
	XMLName xml.Name `xml:"gcm"`
	XMLNS string `xml:"xmlns,attr"`
	Value string `xml:",innerxml"` //A bit of a hack, but it works. Chardata escaped our JSON, but innerxml will not.
}

//BodyStanza stores the data in the body stanza in outgoing messages. Used for marshalling XML.
type BodyStanza struct {
	XMLName xml.Name `xml:"body"`
	Value string `xml:",chardata"`
}

//ACKPayload stores the data in the ACK payload. Used for marshalling JSON.
type ACKPayload struct {
	To string `json:"to"`
	MessageID string `json:"message_id"`
	MessageType string `json:"message_type"`
}

//NewGCMStanza makes a new GCMStanza. the XMLNS should always be google:mobile:data.
func NewGCMStanza(payload string) GCMStanza {
	return GCMStanza {
		Value: payload,
		XMLNS: "google:mobile:data",
	}
}

//NewACKPayload makes a new ACKPayload. MessageType should always be ack.
func NewACKPayload (registrationID, messageID string) ACKPayload {
	return ACKPayload {
			To: registrationID,
			MessageID: messageID,
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
	messageStanza := MessageStanza {
		Body: NewGCMStanza(string(marshaledPayload)),
	}
	marshaledMessageStanza, err := xml.Marshal(messageStanza)
	if err != nil {
		log.Fatal(err)
	}
	return marshaledMessageStanza
}
