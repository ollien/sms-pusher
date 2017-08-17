package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ollien/sms-pusher/server/firebasexmpp"
)

func main() {
	databaseConnection, err := InitDB("./database-config.json")

	if err != nil {
		log.Fatal(err)
	}

	defer databaseConnection.Close()

	supervisor := NewXMPPSupervisor("./xmpp-config.json")
	outChannel := make(chan firebasexmpp.SMSMessage)
	go listenForSMS(databaseConnection, outChannel)
	supervisor.SpawnClient(outChannel)
	fmt.Println("Listening for SMS")
	server := NewWebserver("127.0.0.1:8080", databaseConnection)
	fmt.Println("Server running")
	server.Start()
}

func listenForSMS(databaseConnection *sql.DB, outChannel <-chan firebasexmpp.SMSMessage) {
	for {
		//TODO: Find some way to ping the client of this event. Maybe websockets?
		message := <-outChannel
		fmt.Printf("MESSAGE DETAILS\nFrom: %s\nAt: %d\nBody:%s\n\n", message.PhoneNumber, message.Timestamp, message.Message)
		err := InsertMessage(databaseConnection, message)
		if err != nil {
			log.Fatal(err)
		}
	}
}
