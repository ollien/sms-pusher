package messaging

import (
	"encoding/json"

	"github.com/ollien/sms-pusher/server/firebasexmpp"
)

//ExtractTextMessage will extract a TextMessage from a message sent upstream from FCM
func ExtractTextMessage(message firebasexmpp.UpstreamMessage) (TextMessage, error) {
	mms := MMSMessage{}
	err := json.Unmarshal(message.Data, &mms)
	if err != nil {
		return nil, err
	}

	if mms.isMMS() {
		return mms, nil
	}

	sms := SMSMessage{}
	sms.convertFromMMS(mms)

	return sms, nil
}
