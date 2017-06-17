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
	GCM string `xml:"gcm"`
	XMLNS string `xml:"xmlns,attr"`
}

type BodyStanza struct {
	Body string `xml:"body"`
}

type ACKPayload struct {
	To string `json:"to"`
	MessageId string `json:"message_id"`
	MessageType string `json:"message_type"`
}

func NewGCMStanza(payload string) GCMStanza {
	return GCMStanza {
		GCM: payload,
		XMLNS: "google:mobile:data",
	}
}

func NewACKPayload (registrationId, messageId string){
	return ACKPayload {
			To: registrationId,
			MessageId: messageId,
			messageType: "ack",
		}
}

func ConstructACK(registrationId, messageId string) []byte {
	payload := NewACKPayload(registrationId, messageId)
	marshaledPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	messageStanza := MessageStanza {
		Body: NewGCMStanza(payload),
	}
	marshalledMessageStanza, err := xml.Marshal(messageStanza)
	if err != nil {
		log.Fatal(err)
	}
	return marshalledMessageStanza
}
