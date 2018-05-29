package messaging

import "encoding/json"

//StringEncodedStringSlice represents an array of strings that is encoded as JSON
type StringEncodedStringSlice []string

//TextMessage represents either a SMS or an MMS.
type TextMessage interface {
	isMMS() bool
}

//SMSMessage stores the data sent by the app upstream about incoming SMS messages.
type SMSMessage struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message,omitempty"`
	Timestamp   int64  `json:"timestamp,string"`
}

//MMSMessage represents an MMS message that comes in
type MMSMessage struct {
	SMSMessage
	Recipients  StringEncodedStringSlice `json:"recipients"`
	PartBlockID string                   `json:"block_id"`
}

func (message SMSMessage) isMMS() bool {
	return false
}

func (message *SMSMessage) convertFromMMS(mms MMSMessage) {
	message.PhoneNumber = mms.PhoneNumber
	message.Timestamp = mms.Timestamp
	message.Message = mms.Message
}

func (message MMSMessage) isMMS() bool {
	return len(message.Recipients) > 1 || message.PartBlockID != ""
}

//UnmarshalJSON allows StringEncodedString slice to implement the Unmarshaler interface
func (encodedSlice *StringEncodedStringSlice) UnmarshalJSON(data []byte) error {
	var decodedString string
	err := json.Unmarshal(data, &decodedString)
	if err != nil {
		return err
	}
	stringSlice := (*[]string)(encodedSlice)

	return json.Unmarshal([]byte(decodedString), stringSlice)
}
