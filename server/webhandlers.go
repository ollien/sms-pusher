package main

import (
	"database/sql"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//RouteHandler holds all routes and allows them to share common variables
type RouteHandler struct {
	databaseConnection *sql.DB
}

func (handler RouteHandler) index(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}

func (handler RouteHandler) register(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	username := req.FormValue("username")
	password := req.FormValue("password")

	if username == "" || password == "" || len(password) < 8 {
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encodedPassword := []byte(password)
	err := CreateUser(handler.databaseConnection, username, encodedPassword)

	if err != nil {
		//Postgres specific check
		if err.Error() == duplicateUserError {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
		}

	}
}

func (handler RouteHandler) authenticate(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	cookie := GetSessionCookie(req)
	if cookie != nil {
		_, err := GetUserFromSession(handler.databaseConnection, cookie.Value)
		if err == nil {
			//user exists and session is valid - write 200 and move on
			return
		}
	}
	username := req.FormValue("username")
	password := req.FormValue("password")

	if username == "" || password == "" {
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encodedPassword := []byte(password)
	user, err := VerifyUser(handler.databaseConnection, username, encodedPassword)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessionID, err := CreateSession(handler.databaseConnection, user)
	if err != nil {
		//TODO: Log data about 500
		writer.WriteHeader(http.StatusInternalServerError)
	}

	cookie = &http.Cookie{
		Name:  "session",
		Value: sessionID,
	}
	http.SetCookie(writer, cookie)
}
