package db

import (
	"database/sql"
	"net"

	_ "github.com/lib/pq"
	"github.com/ollien/sms-pusher/server/config"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const driver = "postgres"

//DatabaseConnection represents a single connection to the database
type DatabaseConnection struct {
	*sql.DB
	logger *logrus.Logger
}

//DatabaseError represents an error that was produced during the running of a databse action
//Can either contain an error or a string. If both are present in the struct, the string will take precedence.
type DatabaseError struct {
	err     error
	message string
	//DatabaseFault signals whether or not this is something that should be thought of as an error with the database itself, and not a problem with the query.
	DatabaseFault bool
}

//User represents a user within the database
type User struct {
	ID           int
	Username     string
	passwordHash []byte
}

//Device represents a user within the database
type Device struct {
	ID       int
	DeviceID uuid.UUID
	FCMID    []byte
	User     User
}

//Session represents a session for a user
type Session struct {
	ID   string
	User User
}

//InitDB intiializes the database connection and returns a DB
func InitDB(logger *logrus.Logger) (DatabaseConnection, error) {
	appConfig, err := config.GetConfig()
	if err != nil {
		return DatabaseConnection{}, err
	}

	rawConnection, err := sql.Open(driver, appConfig.Database.URI)
	if err != nil {
		return DatabaseConnection{}, err
	}

	connection := DatabaseConnection{
		DB:     rawConnection,
		logger: logger,
	}

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

	//Create mms_file_blocks table
	_, err = connection.Exec("CREATE TABLE IF NOT EXISTS mms_file_blocks(" +
		"id uuid PRIMARY KEY," +
		"for_user INTEGER REFERENCES users(id));")
	if err != nil {
		return DatabaseConnection{}, err
	}

	//Create mms_files table
	//name is 128 bytes, as the file hash will be 64 bytes, plus we need room for the extension. Rounded up to the nearest power of two, 128.
	_, err = connection.Exec("CREATE TABLE IF NOT EXISTS mms_files(" +
		"id SERIAL PRIMARY KEY," +
		"name VARCHAR(128)," +
		"block uuid REFERENCES mms_file_blocks(id));")
	if err != nil {
		return DatabaseConnection{}, err
	}

	return connection, nil
}

//handleError will take an error, package it as a DatabaseError, and perform any logging needed.
func (connection DatabaseConnection) handleError(err error, databaseFault bool) DatabaseError {
	logged := connection.logIfNetError(err)
	return DatabaseError{
		err:           err,
		DatabaseFault: databaseFault || logged,
	}
}

//logIfNetError will log the error with error severity if the error spawned from a network problem, such as the database being down. Returns true if an error was logged.
func (connection DatabaseConnection) logIfNetError(err error) bool {
	if _, ok := err.(*net.OpError); ok {
		connection.logger.WithField("err", err).Error("Could not connect to database.")
		return true
	}

	return false
}

//Error returns the error message associated with a database error.
//If a string and error are present in the DatabaseError, the string is returned.
//Allows it to implement the error interface.
func (err DatabaseError) Error() string {
	if err.message == "" {
		return err.err.Error()
	}

	return err.message
}
