package main

import (
	"database/sql"
	"errors"
	"net/http"
)

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
func GetSessionUser(db *sql.DB, req *http.Request) (User, error) {
	cookie := GetSessionCookie(req)
	sessionID := req.FormValue("session_id")
	if sessionID == "" {
		if cookie != nil {
			sessionID = cookie.Value
		} else {
			return User{}, errors.New("no session cookie found")
		}
	}

	user, err := GetUserFromSession(db, sessionID)
	//If err is nil, there is no valid session.
	return user, err

}
