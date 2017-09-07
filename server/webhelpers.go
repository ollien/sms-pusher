package main

import (
	"database/sql"
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

//HasValidSessionCookie checks if the user has a valid session cookie. Helper function for most authentication base
func HasValidSessionCookie(db *sql.DB, req *http.Request) bool {
	cookie := GetSessionCookie(req)
	if cookie != nil {
		_, err := GetUserFromSession(db, cookie.Value)
		//If err is nil, there is no valid session.
		return err == nil
	}

	return false
}
