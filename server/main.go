package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/ollien/sms-pusher/server/web"
)

func main() {
	db, err := InitDb("./database-conf.json")
	if err != nil {
		log.Fatal(err)
	}
	supervisor := NewXMPPSupervisor("./xmpp-config.json")
	outChannel := make(chan firebasexmpp.SMSMessage)
	go listenForSMS(db, outChannel)
	supervisor.SpawnClient(outChannel)
	fmt.Println("Listening for SMS")
	server := web.NewWebserver("127.0.0.1:8080")
	fmt.Println("Server running")
	server.Start()
}

func listenForSMS(db *sql.DB, outChannel <-chan firebasexmpp.SMSMessage) {
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
