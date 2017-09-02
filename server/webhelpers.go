package main

import "net/http"

//GetSessionCookie gets the cookie named "session" from http.Cookies()
func GetSessionCookie(req *http.Request) *http.Cookie {
	for cookie := range req.Cookies() {
		if cookie.Name == "session" {
			return cookie
		}
	}
	return nil
}
