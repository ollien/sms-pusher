package main

import (
	"net/http"
	"time"
)

//Webserver hosts a webserver for sms-pusher
type Webserver struct {
	httpServer *http.Server
	listenAddr string
}

//NewWebserver creats a new Webserver with httpServer being set to a new http.Server
func NewWebserver(listenAddr string) Webserver {
	httpServer := &http.Server{
		Addr:              listenAddr,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	serv := Webserver{
		httpServer: httpServer,
		listenAddr: listenAddr,
	}
	serv.initHandlers()
	return serv
}

func (serv *Webserver) initHandlers() {
	http.HandleFunc("/", index)
}

//Start starts the webserver
func (serv *Webserver) Start() {
	serv.httpServer.ListenAndServe()
}
