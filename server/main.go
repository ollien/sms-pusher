package main

import "fmt"
import "./firebase-xmpp"

func main() {
	client := firebase_xmpp.NewFirebaseClient("./xmpp-config.json")
	outChannel:= client.StartRecv()
	go listenForSMS(outChannel)
	fmt.Println("Listening for SMS")
	server := NewServer("127.0.0.1:8080")
	fmt.Println("Server running")
	server.start()
}

func listenForSMS(outChannel <-chan interface{}) {
	for {
		message := (<-outChannel).(firebase_xmpp.SMSMessage)
		//TODO: Do more than just print the message to terminal. Let's make this project great :)
		fmt.Printf("MESSAGE DETAILS\nFrom: %s\nAt: %d\nBody:%s\n\n", message.PhoneNumber, message.Timestamp, message.Message)
	}
}
