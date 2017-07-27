package main

import (
	"fmt"
	"log"

	"github.com/ollien/sms-pusher/server/core"
	"github.com/ollien/sms-pusher/server/core/firebasexmpp"
)

func main() {
	db, err := core.InitDB("./database-conf.json")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	supervisor := core.NewXMPPSupervisor("./xmpp-config.json")
	outChannel := make(chan firebasexmpp.SMSMessage)
	go core.ListenForSMS(db, outChannel)
	supervisor.SpawnClient(outChannel)
	fmt.Println("Listening for SMS")
	server := NewWebserver("127.0.0.1:8080", db)
	fmt.Println("Server running")
	server.Start()
}
