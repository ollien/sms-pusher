package web

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/sirupsen/logrus"
)

const (
	//1 MB
	maxRequestSize = 4096
	//16 MB
	maxFileSize = 16777216
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
	logger       *logrus.Logger
	router       *Router
	routeHandler RouteHandler
}

type handlerFunction = func(http.ResponseWriter, *http.Request, httprouter.Params)

//NewWebserver creats a new Webserver with httpServer being set to a new http.Server
func NewWebserver(listenAddr string, databaseConnection db.DatabaseConnection, sendChannel chan<- firebasexmpp.OutboundMessage, logger *logrus.Logger) Webserver {
	routeHandler := RouteHandler{
		databaseConnection: databaseConnection,
		sendChannel:        sendChannel,
		logger:             logger,
	}
	serv := Webserver{
		listenAddr:   listenAddr,
		logger:       logger,
		router:       NewRouter(),
		routeHandler: routeHandler,
	}
	serv.router.AfterRequest = serv.afterRequest
	serv.initHandlers()
	return serv
}

func (serv *Webserver) initHandlers() {
	serv.router.GET("/", wrapHandlerFunction(serv.routeHandler.index))
	serv.router.POST("/register", wrapHandlerFunction(serv.routeHandler.register))
	serv.router.POST("/authenticate", wrapHandlerFunction(serv.routeHandler.authenticate))
	serv.router.POST("/register_device", wrapHandlerFunction(serv.routeHandler.registerDevice))
	serv.router.POST("/set_fcm_id", wrapHandlerFunction(serv.routeHandler.setFCMID))
	serv.router.POST("/send_message", wrapHandlerFunction(serv.routeHandler.sendMessage))
}

//wrapHandlerFunction allows us to enforce a file size limit
//Though we could theoretically put this in ServeHTTP, this allows us to set different sizes for different routes after httprouter has taken care of the route handling for us.
func wrapHandlerFunction(handler handlerFunction) handlerFunction {
	return wrapHandlerFunctionWithLimit(handler, maxRequestSize)
}

func wrapHandlerFunctionWithLimit(handler handlerFunction, sizeLimit int64) handlerFunction {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		//Enforce a max file size
		req.Body = http.MaxBytesReader(writer, req.Body, sizeLimit)
		handler(writer, req, params)
	}
}

func (serv *Webserver) afterRequest(writer http.ResponseWriter, req *http.Request) {
	loggableWriter := writer.(*LoggableResponseWriter)
	serv.logger.Infof("%s %s %s %s (%s); %d; %d bytes", req.RemoteAddr, req.Proto, req.Method, req.RequestURI, req.UserAgent(), loggableWriter.statusCode, loggableWriter.bytesWritten)
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
