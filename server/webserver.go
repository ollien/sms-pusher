package main

import (
	"database/sql"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//Webserver hosts a webserver for sms-pusher
type Webserver struct {
	listenAddr   string
	router       *httprouter.Router
	routeHandler RouteHandler
}

//NewWebserver creats a new Webserver with httpServer being set to a new http.Server
func NewWebserver(listenAddr string, database *sql.DB) Webserver {
	routeHandler := RouteHandler{
		database: database,
	}
	serv := Webserver{
		listenAddr:   listenAddr,
		router:       httprouter.New(),
		routeHandler: routeHandler,
	}
	serv.initHandlers()
	return serv
}

func (serv *Webserver) initHandlers() {
	serv.router.GET("/", serv.routeHandler.index)
}

//Start starts the webserver
func (serv *Webserver) Start() {
	http.ListenAndServe(":8080", serv.router)
}
