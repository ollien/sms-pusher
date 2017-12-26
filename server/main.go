package main

import (
	"fmt"
	"log"

	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/ollien/sms-pusher/server/web"
	"github.com/sirupsen/logrus"
)

func main() {
	databaseConnection, err := db.InitDB("./database-config.json")
	if err != nil {
		log.Fatal(err)
	}

	defer databaseConnection.Close()

	supervisor := NewXMPPSupervisor("./xmpp-config.json")
	logger := logrus.New()
	outChannel := make(chan firebasexmpp.SMSMessage)
	sendChannel := make(chan firebasexmpp.OutboundMessage)
	sendErrorChannel := make(chan error)
	go listenForSMS(outChannel)
	supervisor.SpawnClient(outChannel, sendChannel, sendErrorChannel)
	fmt.Println("Listening for SMS")
	server := web.NewWebserver("0.0.0.0:8080", databaseConnection, sendChannel, logger)
	fmt.Println("Server running")
	server.Start()
}

func listenForSMS(outChannel <-chan firebasexmpp.SMSMessage) {
	for {
		//TODO: Find some way to ping the client of this event. Maybe websockets?
		message := <-outChannel
		fmt.Printf("MESSAGE DETAILS\nFrom: %s\nAt: %d\nBody:%s\n\n", message.PhoneNumber, message.Timestamp, message.Message)
	}
}
