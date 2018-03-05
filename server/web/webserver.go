package web

import (
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/sirupsen/logrus"
)

//Router allows for us to direct http requests to the correct handlers with the addition of pre/post request hooks
type Router struct {
	BeforeRequest func(http.ResponseWriter, *http.Request)
	AfterRequest  func(http.ResponseWriter, *http.Request)
	httprouter.Router
}

//Webserver hosts a webserver for sms-pusher
type Webserver struct {
	listenAddr   string
	router       *Router
	routeHandler RouteHandler
}

//NewWebserver creats a new Webserver with httpServer being set to a new http.Server
func NewWebserver(listenAddr string, databaseConnection db.DatabaseConnection, sendChannel chan<- firebasexmpp.OutboundMessage, logger *logrus.Logger) Webserver {
	routeHandler := RouteHandler{
		databaseConnection: databaseConnection,
		sendChannel:        sendChannel,
		logger:             logger,
	}
	serv := Webserver{
		listenAddr:   listenAddr,
		router:       NewRouter(),
		routeHandler: routeHandler,
	}
	serv.router.AfterRequest = serv.afterRequest
	serv.initHandlers()
	return serv
}

func (serv *Webserver) initHandlers() {
	serv.router.GET("/", serv.routeHandler.index)
	serv.router.POST("/register", serv.routeHandler.register)
	serv.router.POST("/authenticate", serv.routeHandler.authenticate)
	serv.router.POST("/register_device", serv.routeHandler.registerDevice)
	serv.router.POST("/set_fcm_id", serv.routeHandler.setFCMID)
	serv.router.POST("/send_message", serv.routeHandler.sendMessage)
}

func (serv *Webserver) afterRequest(writer http.ResponseWriter, req *http.Request) {
	loggableWriter := writer.(*LoggableResponseWriter)
	reqTime := time.Now().Format("2006-01-02 15:04:05-0700")
	serv.routeHandler.logger.Infof("[%s] - %s %s %s %s (%s); %d; %d bytes", reqTime, req.RemoteAddr, req.Proto, req.Method, req.RequestURI, req.UserAgent(), loggableWriter.statusCode, loggableWriter.bytesWritten)
}

//Start starts the webserver
func (serv *Webserver) Start() {
	err := http.ListenAndServe(serv.listenAddr, serv.router)
	if err != nil {
		log.Fatal(err)
	}
}

//NewRouter creates a new Router[:w
func NewRouter() *Router {
	return &Router{
		Router: *httprouter.New(),
	}
}

func (router *Router) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	loggableWriter := NewLoggableResponseWriter(writer)
	if router.BeforeRequest != nil {
		router.BeforeRequest(&loggableWriter, req)
	}
	router.Router.ServeHTTP(&loggableWriter, req)
	if router.AfterRequest != nil {
		router.AfterRequest(&loggableWriter, req)
	}
}
