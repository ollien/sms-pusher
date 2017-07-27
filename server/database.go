package main

import (
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/lib/pq"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"golang.org/x/crypto/bcrypt"
)

const configURIKey = "uri"
const passwordCost = 10

//InitDB intiializes the database connection and returns a DB
func InitDB(configPath string) (*sql.DB, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	jsonDecoder := json.NewDecoder(file)
	configMap := make(map[string]string)
	jsonDecoder.Decode(&configMap)
	databaseConnection, err := sql.Open("postgres", configMap[configURIKey])
	if err != nil {
		return nil, err
	}

	//Create messages table
	_, err = databaseConnection.Exec("CREATE TABLE IF NOT EXISTS messages (" +
		"id SERIAL," +
		"phone_number VARCHAR(16)," +
		"time timestamp," +
		"message TEXT);")

	if err != nil {
		return nil, err
	}

	//Create users table.
	//Our bcrypt implenetation uses 60 char hashes, so we can safely use CHAR(60) as the datatype.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"id SERIAL," +
		"username VARCHAR(32)," +
		"password_hash CHAR(60));")

	//Create devices table
	//UUID4s are 32 hex bits plus four digits by definition, thus we can use a CHAR(36) as the datatype.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS devices (" +
		"id SERIAL," +
		"device_id CHAR(36)," +
		"device_user INTEGER REFERENCES users(id));")

	if err != nil {
		return nil, err
	}

	return databaseConnection, nil
}

//InsertMessage inserts a SMS message into the database
func InsertMessage(databaseConnection *sql.DB, message firebasexmpp.SMSMessage) error {
	_, err := databaseConnection.Exec("INSERT INTO messages VALUES (DEFAULT, $1, to_timestamp($2), $3)", message.PhoneNumber, message.Timestamp, message.Message)
	return err
}

//CreateUser insersts a user into the database
func CreateUser(db *sql.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)

	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users VALUES(DEFAULT, $1, $2)", username, hash)

	return err
}
