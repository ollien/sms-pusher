package firebase_xmpp

import "encoding/json"

type UpstreamMessage struct {
	From string
	TTL int
	MessageId string
	Category string
	Data SMSMessage
}

type SMSMessage struct {
	PhoneNumber string
	Message string
	Timestamp int64
}
