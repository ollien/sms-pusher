package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//Webserver hosts a webserver for sms-pusher
type Webserver struct {
	listenAddr string
	router     *httprouter.Router
}

//NewWebserver creats a new Webserver with httpServer being set to a new http.Server
func NewWebserver(listenAddr string) Webserver {
	serv := Webserver{
		listenAddr: listenAddr,
		router:     httprouter.New(),
	}
	serv.initHandlers()
	return serv
}

func (serv *Webserver) initHandlers() {
	serv.router.GET("/", index)
}

//Start starts the webserver
func (serv *Webserver) Start() {
	http.ListenAndServe(":8080", serv.router)
}
