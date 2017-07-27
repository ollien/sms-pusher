package core

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ollien/sms-pusher/server/core/firebasexmpp"
)

//ListenForSMS listens on outChannel for SMS messages
func ListenForSMS(db *sql.DB, outChannel <-chan firebasexmpp.SMSMessage) {
	for {
		//TODO: Find some way to ping the client of this event. Maybe websockets?
		message := <-outChannel
		fmt.Printf("MESSAGE DETAILS\nFrom: %s\nAt: %d\nBody:%s\n\n", message.PhoneNumber, message.Timestamp, message.Message)
		err := InsertMessage(db, message)
		if err != nil {
			log.Fatal(err)
		}
	}
}
