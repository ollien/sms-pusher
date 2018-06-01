package db

import (
	"database/sql"
	"net"

	_ "github.com/lib/pq"
	"github.com/ollien/sms-pusher/server/config"
	//Adds migrations to goose
	_ "github.com/ollien/sms-pusher/server/db/migrations"
	"github.com/pressly/goose"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const driver = "postgres"

//DatabaseConnection represents a single connection to the database
type DatabaseConnection struct {
	*sql.DB
	logger *logrus.Logger
	uri    string
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
	ID    uuid.UUID
	FCMID []byte
	User  User
}

//Session represents a session for a user
type Session struct {
	ID   uuid.UUID
	User User
}

//NewDatabaseConnection intiializes the database connection and returns a DatabaseConnection.
func NewDatabaseConnection(logger *logrus.Logger) (DatabaseConnection, error) {
	goose.SetLogger(logger)
	appConfig, err := config.GetConfig()
	if err != nil {
		return DatabaseConnection{}, err
	}

	connection := DatabaseConnection{
		logger: logger,
		uri:    appConfig.Database.URI,
	}

	return connection, nil
}

//Connect connets to the database and begins necessary migrations.
func (connection *DatabaseConnection) Connect() error {
	rawConnection, err := sql.Open(driver, connection.uri)
	if err != nil {
		return err
	}

	err = goose.Up(rawConnection, "/dev/null")
	if err != nil {
		return err
	}

	//Don't store the database connection until we know everything's been set up by the migration.
	connection.DB = rawConnection

	return nil
}

//handleError will take an error, package it as a DatabaseError, and perform any logging needed.
func (connection *DatabaseConnection) handleError(err error, databaseFault bool) error {
	if err == nil {
		return nil
	}
	logged := connection.logIfNetError(err)
	return &DatabaseError{
		err:           err,
		DatabaseFault: databaseFault || logged,
	}
}

//logIfNetError will log the error with error severity if the error spawned from a network problem, such as the database being down. Returns true if an error was logged.
func (connection *DatabaseConnection) logIfNetError(err error) bool {
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
