package main

import (
	"database/sql"
	"fmt"
	"log"

	"./firebase-xmpp"
)

func main() {
	db, err := InitDb("./database-conf.json")
	if err != nil {
		log.Fatal(err)
	}
	clients := make(map[string]firebasexmpp.FirebaseClient)
	outChannel:= StartFirebaseClient(clients, "./xmpp-config.json")
	go listenForSMS(db, outChannel)
	fmt.Println("Listening for SMS")
	server := NewServer("127.0.0.1:8080")
	fmt.Println("Server running")
	server.start()
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
