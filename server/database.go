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

//InitDb intiializes the database connection and returns a DB
func InitDb(configPath string) (*sql.DB, error) {
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

	//Create messages table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (" +
		"id SERIAL," +
		"phone_number VARCHAR(16)," +
		"time timestamp," +
		"message TEXT);")

	if err != nil {
		return nil, err
	}

	//Create users table. Our bcrypt implenetation uses 60 char hashes, so we can safely use CHAR(60) as the datatype.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"id SERIAL," +
		"username VARCHAR(32)," +
		"password_hash CHAR(60));")

	return db, nil
}

//InsertMessage inserts a SMS message into the database
func InsertMessage(db *sql.DB, message firebasexmpp.SMSMessage) error {
	_, err := db.Exec("INSERT INTO messages VALUES (DEFAULT, $1, to_timestamp($2), $3)", message.PhoneNumber, message.Timestamp, message.Message)
	return err
}

//CreateUser insersts a user into the database
func CreateUser(db *sql.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)

	if err != nil {
		return err
	}

	_, err := db.Exec("INSERT INTO users VALUES(DEFAULT, $1, $2)", username, hash)

	return err
}
