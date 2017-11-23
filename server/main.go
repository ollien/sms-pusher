package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/ollien/sms-pusher/server/web"
)

func main() {
	databaseConnection, err := db.InitDB("./database-config.json")

	if err != nil {
		log.Fatal(err)
	}

	defer databaseConnection.Close()

	supervisor := NewXMPPSupervisor("./xmpp-config.json")
	outChannel := make(chan firebasexmpp.SMSMessage)
	go listenForSMS(databaseConnection, outChannel)
	supervisor.SpawnClient(outChannel)
	fmt.Println("Listening for SMS")
	server := web.NewWebserver("0.0.0.0:8080", databaseConnection)
	fmt.Println("Server running")
	server.Start()
}

func listenForSMS(databaseConnection *sql.DB, outChannel <-chan firebasexmpp.SMSMessage) {
	for {
		//TODO: Find some way to ping the client of this event. Maybe websockets?
		message := <-outChannel
		fmt.Printf("MESSAGE DETAILS\nFrom: %s\nAt: %d\nBody:%s\n\n", message.PhoneNumber, message.Timestamp, message.Message)
		err := db.InsertMessage(databaseConnection, message)
		if err != nil {
			log.Fatal(err)
		}
	}
}
