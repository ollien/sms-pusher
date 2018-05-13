package main

import (
	"fmt"

	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/ollien/sms-pusher/server/web"
	"github.com/sirupsen/logrus"
)

func main() {
	logFormatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05-0700",
	}
	logger := logrus.New()
	logger.Formatter = logFormatter
	databaseConnection, err := db.InitDB(logger)
	if err != nil {
		logger.Fatalf("Database error: %s", err)
	}

	defer databaseConnection.Close()

	outChannel := make(chan firebasexmpp.TextMessage)
	sendChannel := make(chan firebasexmpp.OutboundMessage)
	supervisor := NewXMPPSupervisor(outChannel, sendChannel, logger)
	go listenForSMS(outChannel)
	supervisor.SpawnClient()
	fmt.Println("Listening for SMS")
	server := web.NewWebserver("0.0.0.0:8080", databaseConnection, sendChannel, logger)
	fmt.Println("Server running")
	server.Start()
}

func listenForSMS(outChannel <-chan firebasexmpp.TextMessage) {
	for {
		//TODO: Find some way to ping the client of this event. Maybe websockets?
		switch message := (<-outChannel).(type) {
		case firebasexmpp.SMSMessage:
			fmt.Printf("MESSAGE DETAILS\nFrom: %s\nAt: %d\nBody:%s\n\n", message.PhoneNumber, message.Timestamp, message.Message)
		case firebasexmpp.MMSMessage:
			fmt.Printf("MESSAGE DETAILS\nFrom: %s\nTo:%v\nAt: %d\nBody:%s\nPartsBlockID:%s\n\n", message.PhoneNumber, message.Recipients, message.Timestamp, message.Message, message.PartBlockID)
		}
	}
}
