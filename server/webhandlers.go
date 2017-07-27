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

	CreateUser(handler.database, username, password)
}
