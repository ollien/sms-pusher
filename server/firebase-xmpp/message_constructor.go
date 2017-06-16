package firebase_xmpp

import "encoding/xml"

type XMPPMessage struct {
	XMLName xml.Name `xml:"message"`
	To string `xml:"to,attr"`
	Type string `xml:"type,attr"`
	Id string `xml:"id,attr"`
	Body string `xml:"body"`
}
