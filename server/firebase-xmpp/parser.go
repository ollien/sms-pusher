package firebase_xmpp

type UpstreamMessage struct {
	From string `json:"from"`
	TTL int `json:"time_to_live"`
	MessageId string `json:"message_id"`
	Category string `json:"category"`
	Data SMSMessage
}

type InboundACKMessage struct {
	From string `json:"from"`
	MessageId string `json:"message_id"`
}

type NACKMessage struct {
	From string `json:"from"`
	MessageId string `json:"message_id"`
	Error string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type SMSMessage struct {
	PhoneNumber string `json:"phone_number"`
	Message string `json:"message"`
	Timestamp int64 `json:"timestamp,string"`
}

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
