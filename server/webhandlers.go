package main

import (
	"database/sql"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//RouteHandler holds all routes and allows them to share common variables
type RouteHandler struct {
	database *sql.DB
}

func (handler RouteHandler) index(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}
