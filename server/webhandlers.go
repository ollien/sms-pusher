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
	username := req.FormValue("username")
	password := req.FormValue("password")

	if username == "" || password == "" {
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
	}

	encodedPassword := []byte(password)
	verified, err := VerifyUser(handler.databaseConnection, username, encodedPassword)
	if err != nil {
		//TODO: Log data about 500
		writer.WriteHeader(http.StatusInternalServerError)
	} else if !verified {
		writer.WriteHeader(http.StatusUnauthorized)
	}
	//TODO: write cookies with sessions and handle them for authentication
}
