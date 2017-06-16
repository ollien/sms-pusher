package firebase_xmpp

import "strconv"
import "encoding/json"

type UpstreamMessage struct {
	From string
	TTL int
	MessageId string
	Category string
	Data SMSMessage
}

type InboundACKMessage struct {
	To string
	MessageId string
	MessageType string
}

type OutboundACKMessage struct {
	From string
	MessageId string
	MessageType string
}

type NACKMessage struct {
	From string
	MessageId string
	MessageType string
	Error string
	ErrorDescription string
}

type SMSMessage struct {
	PhoneNumber string
	Message string
	Timestamp int64
}

func (message *OutboundACKMessage) UnmarshalJSON(rawData []byte) error {
	messageMap := make(map[string]*json.RawMessage)
	json.Unmarshal(rawData, &messageMap)
	json.Unmarshal(*messageMap["from"], &message.From)
	json.Unmarshal(*messageMap["message_id"], &message.MessageId)
	json.Unmarshal(*messageMap["message_type"], &message.MessageType)
	return nil
}

func (message *NACKMessage) UnmarshalJSON(rawData []byte) error {
	messageMap := make(map[string]*json.RawMessage)
	json.Unmarshal(rawdata, &messageMap)
	json.Unmarshal(*messageMap["from"], &message.From)
	json.Unmarshal(*messageMap["message_id"], &message.MessageId)
	json.Unmarshal(*messageMap["message_type"], &message.MessageType)
	json.Unmarshal(*messageMap["error"], &message.Error)
	json.Unmarshal(*messageMap["error_description"], &message.ErrorDescription)
	return nil
}

func (message *UpstreamMessage) UnmarshalJSON(rawData []byte) error {
	messageMap := make(map[string]*json.RawMessage)
	dataMap := make(map[string]*json.RawMessage)
	json.Unmarshal(rawData, &messageMap)
	message.Data = SMSMessage{}
	json.Unmarshal(*messageMap["from"], &message.From)
	json.Unmarshal(*messageMap["time_to_live"], &message.TTL)
	json.Unmarshal(*messageMap["message_id"], &message.MessageId)
	json.Unmarshal(*messageMap["category"], &message.Category)
	json.Unmarshal(*messageMap["data"], &dataMap)
	json.Unmarshal(*dataMap["phone_number"], &message.Data.PhoneNumber)
	json.Unmarshal(*dataMap["message"], &message.Data.Message)
	//Android only allows us to send strings upstream. In light of this, we must convert the timestamp to int64 before storing it
	var timestamp string
	json.Unmarshal(*dataMap["timestamp"], &timestamp)
	convertedTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return err
	}
	message.Data.Timestamp = convertedTimestamp
	return nil
}
