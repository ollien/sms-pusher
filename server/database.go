package main

import (
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/lib/pq"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	uuid "github.com/satori/go.uuid"
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

//Device represents a user within the database
type Device struct {
	ID         int
	DeviceID   []byte
	DeviceUser User
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
		"id SERIAL PRIMARY KEY," +
		"phone_number VARCHAR(16)," +
		"time timestamp," +
		"message TEXT);")

	if err != nil {
		return nil, err
	}

	//Create users table.
	_, err = databaseConnection.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"id SERIAL PRIMARY KEY," +
		"username VARCHAR(32) UNIQUE," +
		"password_hash bytea);")

	if err != nil {
		return nil, err
	}

	//Create devices table
	_, err = databaseConnection.Exec("CREATE TABLE IF NOT EXISTS devices (" +
		"id SERIAL PRIMARY KEY," +
		"device_id uuid UNIQUE," +
		"device_user INTEGER REFERENCES users(id));")

	if err != nil {
		return nil, err
	}

	//Create sessions table
	_, err = databaseConnection.Exec("CREATE TABLE IF NOT EXISTS sessions (" +
		"id uuid PRIMARY KEY," +
		"for_user INTEGER REFERENCES users(id));")

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
	hash, err := bcrypt.GenerateFromPassword(password, passwordCost)

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
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:           id,
		Username:     internalUsername,
		passwordHash: passwordHash,
	}

	return user, nil
}

//GetUserByID gets a user from the database and returns a User.
func GetUserByID(databaseConnection *sql.DB, id int) (User, error) {
	userRow := databaseConnection.QueryRow("SELECT * FROM users WHERE id = $1", id)
	var internalID int
	var username string
	var passwordHash []byte
	err := userRow.Scan(&internalID, &username, &passwordHash)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:           internalID,
		Username:     username,
		passwordHash: passwordHash,
	}

	return user, nil
}

//VerifyUser verifies a user against its authentication details. Returns the user if authed.
func VerifyUser(databaseConnection *sql.DB, username string, password []byte) (User, error) {
	user, err := GetUser(databaseConnection, username)
	if err != nil {
		return User{}, err
	}
	err = bcrypt.CompareHashAndPassword(user.passwordHash, password)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

//CreateSession makes a session given a User
func CreateSession(databaseConnection *sql.DB, user User) (string, error) {
	sessionID := uuid.NewV4().String()
	_, err := databaseConnection.Exec("INSERT INTO sessions VALUES($1, $2);", sessionID, user.ID)

	return sessionID, err
}

//GetUserFromSession gets the user associated with a session
func GetUserFromSession(databaseConnection *sql.DB, sessionID string) (User, error) {
	sessionRow := databaseConnection.QueryRow("SELECT for_user FROM sessions WHERE id = $1", sessionID)
	var userID int
	err := sessionRow.Scan(&userID)
	if err != nil {
		return User{}, err
	}

	user, err := GetUserByID(databaseConnection, userID)
	if err != nil {
		return User{}, err
	}

	return user, err
}

//GetDevice gets a Device from the database, given a deviceID
func GetDevice(databaseConnection *sql.DB, deviceID []byte) (Device, error) {
	deviceRow := databaseConnection.QueryRow("SELECT * FROM devices WHERE device_id = $1", deviceID)
	var id int
	var internalDeviceID []byte
	var userID int
	err := deviceRow.Scan(&id, &internalDeviceID, &userID)
	if err != nil {
		return Device{}, err
	}

	user, err := GetUserByID(databaseConnection, userID)
	//If there's an error, there's an invalid user for the device. (i.e. doesn't exist)
	if err != nil {
		return Device{}, err
	}

	return Device{
		ID:         id,
		DeviceID:   deviceID,
		DeviceUser: user,
	}, nil

}

//RegisterDeviceToUser registers a device for a user
func RegisterDeviceToUser(databaseConnection *sql.DB, user User) (Device, error) {
	deviceID := uuid.NewV4().String()
	deviceRow := databaseConnection.QueryRow("INSERT INTO devices VALUES(DEFAULT, $1, $2) RETURNING *;", deviceID, user.ID)

	var id int
	var internalDeviceID []byte
	var userID int
	err := deviceRow.Scan(&id, &internalDeviceID, &userID)
	if err != nil {
		return Device{}, err
	}

	return Device{
		ID:         id,
		DeviceID:   internalDeviceID,
		DeviceUser: user,
	}, nil
}
