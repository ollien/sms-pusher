package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/ollien/sms-pusher/server/web"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
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

//setup sets up datbase rows such that the server can function.
//Returns true if any stateful actions were performed.  //Presently, it registers a user to the database if necessary.
func setup(databaseConnection db.DatabaseConnection) (bool, error) {
	//TODO: If this function gets any bigger, refactor it into its own package.
	numUsers, err := databaseConnection.GetUserCount()
	if err != nil {
		return false, err
	}

	if numUsers == 0 {
		fmt.Println("No user registered!")
		fmt.Println("Register a user...")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Username: ")
		rawUsername, err := reader.ReadString('\n')
		//Initialize the passwords to different lengths so they're inherently unequal
		password := make([]byte, 0)
		confirmedPassword := make([]byte, 1)
		//Input a password until both entered passwords are equal
		for string(password) != string(confirmedPassword) {
			fmt.Print("Password: ")
			password, err = terminal.ReadPassword(syscall.Stdin)
			if err != nil {
				return false, err
			}
			fmt.Print("\n")
			fmt.Print("Confirm: ")
			confirmedPassword, err = terminal.ReadPassword(syscall.Stdin)
			if err != nil {
				return false, err
			}
			fmt.Print("\n\n")
		}

		username := strings.TrimRight(rawUsername, "\r\n")

		err = databaseConnection.CreateUser(username, password)
		if err != nil {
			return false, err
		}

		fmt.Print("\n")

		return true, nil
	}

	return false, nil
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
