package firebase_xmpp

import "encoding/xml"
import "encoding/json"
import "log"

type MessageStanza struct {
	XMLName xml.Name `xml:"message"`
	To string `xml:"to,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Id string `xml:"id,attr,omitempty"`
	Body interface{}
}

type GCMStanza struct {
	XMLName xml.Name `xml:"gcm"`
	XMLNS string `xml:"xmlns,attr"`
	Value string `xml: ",chardata"`
}

type BodyStanza struct {
	XMLName xml.Name `xml:"body"`
	Value string `xml:",chardata"`
}

type ACKPayload struct {
	To string `json:"to"`
	MessageId string `json:"message_id"`
	MessageType string `json:"message_type"`
}

func NewGCMStanza(payload string) GCMStanza {
	return GCMStanza {
		Value: payload,
		XMLNS: "google:mobile:data",
	}
}

func NewACKPayload (registrationId, messageId string) ACKPayload {
	return ACKPayload {
			To: registrationId,
			MessageId: messageId,
			MessageType: "ack",
		}
}

func ConstructACK(registrationId, messageId string) []byte {
	payload := NewACKPayload(registrationId, messageId)
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
