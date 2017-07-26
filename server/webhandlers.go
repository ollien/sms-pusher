package main

import (
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func index(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}
