package main

import (
	"io"
	"net/http"
	"time"
)

type Webserver struct {
	http_server *http.Server
	listen_addr string
}

func index(writer http.ResponseWriter, req *http.Request) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}

func NewWebserver(listen_addr string) Webserver {
	http_server := &http.Server {
		Addr: listen_addr,
		ReadTimeout: 5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	serv := Webserver {
		http_server: http_server,
		listen_addr: listen_addr,
	}
	serv.initHandlers()
	return serv
}

func (serv *Webserver) initHandlers() {
	http.HandleFunc("/", index)
}

func (serv *Webserver) start() {
	serv.http_server.ListenAndServe()
}
