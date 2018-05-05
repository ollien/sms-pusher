package web

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/h2non/filetype"
	"github.com/ollien/sms-pusher/server/db"
	"github.com/sirupsen/logrus"
)

const (
	uploadedFileMode = 0644
	routeKey         = "_route"
)

//LoggableResponseWriter allows access to otherwise hidden information within ResponseWriter
type LoggableResponseWriter struct {
	http.ResponseWriter
	statusCode     int
	bytesWritten   int
	headersWritten bool
}

//GetSessionCookie gets the cookie named "session" from http.Cookies()
func GetSessionCookie(req *http.Request) *http.Cookie {
	for _, cookie := range req.Cookies() {
		if cookie.Name == "session" {
			return cookie
		}
	}

	return nil
}

//GetSessionUser gets the user associated with a session within a *http.Request.
func GetSessionUser(databaseConnection db.DatabaseConnection, req *http.Request) (db.User, error) {
	cookie := GetSessionCookie(req)
	sessionID := req.FormValue("session_id")
	if sessionID == "" {
		if cookie != nil {
			sessionID = cookie.Value
		} else {
			return db.User{}, errors.New("no session cookie found")
		}
	}

	session, err := databaseConnection.GetSession(sessionID)

	//If err is not nil, there is no valid session.
	return session.User, err
}

//logWithRoute returns a logrus.Entry that contains a field of the route that is being logged
func logWithRoute(logger *logrus.Logger, req *http.Request) *logrus.Entry {
	return logger.WithField(routeKey, req.RequestURI)
}

//logWithRouteField is equivalent to logrus.WithField, but inserts information about the route that is being logged.
func logWithRouteField(logger *logrus.Logger, req *http.Request, key string, value interface{}) *logrus.Entry {
	fields := make(logrus.Fields)
	fields[key] = value
	return logWithRouteFields(logger, req, fields)
}

//logWithRouteFields is equivalent to logrus.WithFields, but inserts information about the route that is being logged.
func logWithRouteFields(logger *logrus.Logger, req *http.Request, fields logrus.Fields) *logrus.Entry {
	fields[routeKey] = req.RequestURI
	return logger.WithFields(fields)
}

//NewLoggableResponseWriter creats a LoggableResponseWriter with the given http.ResponseWriter
func NewLoggableResponseWriter(writer http.ResponseWriter) LoggableResponseWriter {
	return LoggableResponseWriter{
		ResponseWriter: writer,
	}
}

//StoreFile stores an incoming file to disk, with its SHA256 as its username
func StoreFile(bytes []byte, uploadLocation string) (int, error) {
	fileName, err := getFileName(bytes)
	if err != nil {
		return 0, err
	}
	//TODO: store this path in the database

	filePath := path.Join(uploadLocation, fileName)
	file, err := os.OpenFile(filePath, os.O_WRONLY, uploadedFileMode)
	if err != nil {
		return 0, err
	}

	return file.Write(bytes)
}

//Get the name for a file that we will be storing
func getFileName(bytes []byte) (string, error) {
	fileHash := sha256.Sum256(bytes)
	theType, err := filetype.Match(bytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%x", theType, fileHash), nil
}

//Write is identical to http.ResponseWriter.Write, but stores the bytes sent and accounts for the 200 special case that Write normally handles by the interface's definition.
func (writer *LoggableResponseWriter) Write(bytes []byte) (int, error) {
	//Normally, this would be done by ResponseWriter.Write, but the wrapped response writer will not call this method, thus we have to force this same behavior.
	if !writer.headersWritten {
		writer.WriteHeader(http.StatusOK)
	}

	n, err := writer.ResponseWriter.Write(bytes)
	writer.bytesWritten += n

	return n, err
}

//WriteHeader is identical to http.Responsewriter.WriteHeader, but stores the status code.
func (writer *LoggableResponseWriter) WriteHeader(statusCode int) {
	if !writer.headersWritten {
		writer.headersWritten = true
		writer.statusCode = statusCode
		writer.ResponseWriter.WriteHeader(statusCode)
	}
}
