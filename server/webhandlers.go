package main

import (
	"io"
	"net/http"
)

func index(writer http.ResponseWriter, req *http.Request) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}
