package db

import (
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

const (
	configURIKey = "uri"
	driver       = "postgres"
)

//DatabaseConnection represents a single connection to the database
type DatabaseConnection struct {
	*sql.DB
}

//User represents a user within the database
type User struct {
	ID           int
	Username     string
	passwordHash []byte
}

//Device represents a user within the database
type Device struct {
	ID         int
	DeviceID   uuid.UUID
	FCMID      []byte
	DeviceUser User
}

//Session represents a session for a user
type Session struct {
	ID   string
	User User
}

//InitDB intiializes the database connection and returns a DB
func InitDB(configPath string) (DatabaseConnection, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return DatabaseConnection{}, err
	}

	jsonDecoder := json.NewDecoder(file)
	configMap := make(map[string]string)
	jsonDecoder.Decode(&configMap)
	rawConnection, err := sql.Open(driver, configMap[configURIKey])
	if err != nil {
		return DatabaseConnection{}, err
	}

	connection := DatabaseConnection{rawConnection}

	//Create users table.
	_, err = connection.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"id SERIAL PRIMARY KEY," +
		"username VARCHAR(32) UNIQUE," +
		"password_hash bytea);")
	if err != nil {
		return DatabaseConnection{}, err
	}

	//Create devices table
	_, err = connection.Exec("CREATE TABLE IF NOT EXISTS devices (" +
		"id SERIAL PRIMARY KEY," +
		"device_id uuid UNIQUE," +
		"firebase_id bytea," +
		"for_user INTEGER REFERENCES users(id));")
	if err != nil {
		return DatabaseConnection{}, err
	}

	//Create sessions table
	_, err = connection.Exec("CREATE TABLE IF NOT EXISTS sessions (" +
		"id uuid PRIMARY KEY," +
		"for_user INTEGER REFERENCES users(id));")
	if err != nil {
		return DatabaseConnection{}, err
	}

	return connection, nil
}
