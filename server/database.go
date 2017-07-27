package main

import (
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/lib/pq"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
)

const configURIKey = "uri"

//InitDB intiializes the database connection and returns a DB
func InitDB(configPath string) (*sql.DB, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	jsonDecoder := json.NewDecoder(file)
	configMap := make(map[string]string)
	jsonDecoder.Decode(&configMap)
	db, err := sql.Open("postgres", configMap[configURIKey])
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

//InsertMessage inserts a SMS message into the database
func InsertMessage(db *sql.DB, message firebasexmpp.SMSMessage) error {
	_, err := db.Exec("INSERT INTO messages VALUES (DEFAULT, $1, to_timestamp($2), $3)", message.PhoneNumber, message.Timestamp, message.Message)
	return err
}
