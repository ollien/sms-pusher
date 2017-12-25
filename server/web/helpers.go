package web

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/ollien/sms-pusher/server/db"
	"github.com/sirupsen/logrus"
)

const routeKey = "_route"

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
func GetSessionUser(databaseConnection *sql.DB, req *http.Request) (db.User, error) {
	cookie := GetSessionCookie(req)
	sessionID := req.FormValue("session_id")
	if sessionID == "" {
		if cookie != nil {
			sessionID = cookie.Value
		} else {
			return db.User{}, errors.New("no session cookie found")
		}
	}

	user, err := db.GetUserFromSession(databaseConnection, sessionID)

	//If err is not nil, there is no valid session.
	return user, err
}

//LogWithRouteField is equivalent to logrus.WithField, but inserts information about the route that is being logged.
func LogWithRouteField(logger *logrus.Logger, req *http.Request, key string, value interface{}) *logrus.Entry {
	fields := make(logrus.Fields)
	fields[key] = value
	return LogWithRouteFields(logger, req, fields)
}

//LogWithRouteFields is equivalent to logrus.WithFields, but inserts information about the route that is being logged.
func LogWithRouteFields(logger *logrus.Logger, req *http.Request, fields logrus.Fields) *logrus.Entry {
	fields[routeKey] = req.RequestURI
	return logger.WithFields(fields)
}
