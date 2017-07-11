package firebase_xmpp

//UpstreamMessage stores the basic data from any upstream Firebase Cloud Messaging XML Message.
//This isn't as general as it could be. Because the app only sends SMS messages upstream, I've included an SMSMessage in UpstreaMessage.
type UpstreamMessage struct {
	From string `json:"from"`
	TTL int `json:"time_to_live"`
	MessageID string `json:"message_id"`
	Category string `json:"category"`
	Data SMSMessage
}

//InboundACKMessage stores the basic data from an ACK message that Firebase CLoud Messaging when we send a message downstream.
type InboundACKMessage struct {
	From string `json:"from"`
	MessageID string `json:"message_id"`
}

//NACKMessage stores the basic data from an NACK message that Firebase CLoud Messaging when we send a message downstream.
type NACKMessage struct {
	From string `json:"from"`
	MessageID string `json:"message_id"`
	Error string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

//SMSMessage stores the data sent by the app upstream about incoming SMS messages.
type SMSMessage struct {
	PhoneNumber string `json:"phone_number"`
	Message string `json:"message"`
	Timestamp int64 `json:"timestamp,string"`
}

//GetMessageType determines the type of message that Firebase Cloud Messaging has sent upstream.
func GetMessageType(rawData[] byte) string {
	dataMap := make(map[string]string)
	if value, exists := dataMap["type"]; exists {
		if value == "ack" {
			return "InboundACKMessage"
		} else if value == "nack" {
			return "NACKMessage"
		}
	}
	return "UpstreamMessage"
}
