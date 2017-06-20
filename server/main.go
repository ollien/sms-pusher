package main

import "database/sql"
import "fmt"
import "./firebase-xmpp"
import "log"

func main() {
	db, err := initDb("./database-conf.json")
	if err != nil {
		log.Fatal(err)
	}
	client := firebase_xmpp.NewFirebaseClient("./xmpp-config.json")
	outChannel:= client.StartRecv()
	go listenForSMS(db, outChannel)
	fmt.Println("Listening for SMS")
	server := NewServer("127.0.0.1:8080")
	fmt.Println("Server running")
	server.start()
}

func listenForSMS(db *sql.DB, outChannel <-chan interface{}) {
	for {
		//TODO: Find some way to ping the client of this event. Maybe websockets?
		message := (<-outChannel).(firebase_xmpp.SMSMessage)
		fmt.Printf("MESSAGE DETAILS\nFrom: %s\nAt: %d\nBody:%s\n\n", message.PhoneNumber, message.Timestamp, message.Message)
		err := insertMessage(db, message)
		if err != nil {
			log.Fatal(err)
		}
	}
}
