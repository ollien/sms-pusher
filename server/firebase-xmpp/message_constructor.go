package firebase_xmpp

import "encoding/xml"

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
