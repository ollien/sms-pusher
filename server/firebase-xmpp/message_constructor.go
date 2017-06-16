package firebase_xmpp

import "encoding/xml"

type MessageStanza struct {
	XMLName xml.Name `xml:"message"`
	To string `xml:"to,attr"`
	Type string `xml:"type,attr"`
	Id string `xml:"id,attr"`
	Body interface{}
}

type GCMStanza struct {
	GCM string `xml:"gcm"`
	XMLNS string `xml:"xmlns,attr"`
}

type BodyStanza struct {
	Body string `xml:"body"`
}
