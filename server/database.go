package main

import "encoding/json"
//import "fmt"
import "database/sql"
import _ "github.com/lib/pq"
import "./firebase-xmpp"
import "os"

const CONFIG_URI_KEY = "uri"

func InitDb(configPath string) (*sql.DB, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	jsonDecoder := json.NewDecoder(file)
	configMap := make(map[string]string)
	jsonDecoder.Decode(&configMap)
	db, err := sql.Open("postgres", configMap[CONFIG_URI_KEY])
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (" +
			"id SERIAL," +
			"phone_number VARCHAR(16)," +
			"time timestamp," +
			"message TEXT);")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func InsertMessage(db *sql.DB, message firebase_xmpp.SMSMessage) error {
	_, err := db.Exec("INSERT INTO messages VALUES (DEFAULT, $1, to_timestamp($2), $3)", message.PhoneNumber, message.Timestamp, message.Message)
	return err
}

func FindDuplicate(db *sql.DB, message firebase_xmpp.SMSMessage) (firebase_xmpp.SMSMessage, error) {
	rows, err := db.Query("SELECT * FROM messages WHERE phone_number = $1, time = to_timestamp($2), mesage = $3", message.PhoneNumber, message.Timestamp, message.Message)
}
