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

//Error codes for checking
const duplicateUserError = "pq: duplicate key value violates unique constraint \"users_username_key\""

//User represents a user within the database
type User struct {
	ID           int
	Username     string
	passwordHash []byte
}

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
	_, err = databaseConnection.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"id SERIAL," +
		"username VARCHAR(32) UNIQUE," +
		"password_hash CHAR(60));")

	//Create devices table
	//UUID4s are 32 hex bits plus four digits by definition, thus we can use a CHAR(36) as the datatype.
	_, err = databaseConnection.Exec("CREATE TABLE IF NOT EXISTS devices (" +
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
func CreateUser(databaseConnection *sql.DB, username string, password []byte) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)

	if err != nil {
		return err
	}

	_, err = databaseConnection.Exec("INSERT INTO users VALUES(DEFAULT, $1, $2)", username, hash)

	return err
}

//GetUser gets a user from the database and returns a User.
func GetUser(databaseConnection *sql.DB, username string) (User, error) {
	userRow := databaseConnection.QueryRow("SELECT * FROM users WHERE username = $1", username)
	var id int
	var internalUsername string
	var passwordHash []byte
	err := userRow.Scan(&id, &internalUsername, &passwordHash)
	user := User{
		ID:           id,
		Username:     internalUsername,
		passwordHash: passwordHash,
	}
	return user, err
}

//VerifyUser verifies a user against its authentication details. Returns true if authed.
func VerifyUser(databaseConnection *sql.DB, username string, password []byte) (bool, error) {
	user, err := GetUser(databaseConnection, username)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword(user.passwordHash, password)
	if err != nil {
		return false, err
	}
	return true, nil
}
