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
func NewWebserver(listenAddr string, databaseConnection *sql.DB) Webserver {
	routeHandler := RouteHandler{
		databaseConnection: databaseConnection,
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
	serv.router.POST("/register", serv.routeHandler.register)
	serv.router.POST("/authenticate", serv.routeHandler.authenticate)
	serv.router.POST("/register_device", serv.routeHandler.registerDevice)
	serv.router.POST("/set_fcm_id", serv.routeHandler.setFCMID)
}

//Start starts the webserver
func (serv *Webserver) Start() {
	http.ListenAndServe(serv.listenAddr, serv.router)
}
