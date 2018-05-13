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

//Webserver hosts a webserver for sms-pusher
type Webserver struct {
	listenAddr   string
	router       *httprouter.Router
	routeHandler RouteHandler
}

type handlerFunction = func(http.ResponseWriter, *http.Request, httprouter.Params)

//NewWebserver creats a new Webserver with httpServer being set to a new http.Server
func NewWebserver(listenAddr string, databaseConnection db.DatabaseConnection, sendChannel chan<- firebasexmpp.OutboundMessage, logger *logrus.Logger) Webserver {
	routeHandler := RouteHandler{
		databaseConnection: databaseConnection,
		sendChannel:        sendChannel,
		logger:             newRouteLogger(logger),
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
	serv.router.GET("/", serv.wrapHandlerFunction(serv.routeHandler.index))
	serv.router.POST("/register", serv.wrapHandlerFunction(serv.routeHandler.register))
	serv.router.POST("/authenticate", serv.wrapHandlerFunction(serv.routeHandler.authenticate))
	serv.router.POST("/register_device", serv.wrapHandlerFunction(serv.routeHandler.registerDevice))
	serv.router.POST("/set_fcm_id", serv.wrapHandlerFunction(serv.routeHandler.setFCMID))
	serv.router.POST("/send_message", serv.wrapHandlerFunction(serv.routeHandler.sendMessage))
	serv.router.POST("/upload_mms_file", serv.wrapHandlerFunctionWithLimit(serv.routeHandler.uploadMMSFile, maxFileSize))
}

//wrapHandlerFunction allows us to enforce a file size limit
//Though we could theoretically put this in ServeHTTP, this allows us to set different sizes for different routes after httprouter has taken care of the route handling for us.
func (serv *Webserver) wrapHandlerFunction(handler handlerFunction) handlerFunction {
	return serv.wrapHandlerFunctionWithLimit(handler, maxRequestSize)
}

//wrapHandlerFunctionWithLimit is the same as wrapHandlerFunction but allows us to set a size limit on the request
func (serv *Webserver) wrapHandlerFunctionWithLimit(handler handlerFunction, sizeLimit int64) handlerFunction {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		loggableWriter := NewLoggableResponseWriter(writer)
		//Enforce a max file size
		req.Body = http.MaxBytesReader(&loggableWriter, req.Body, sizeLimit)
		//Pass our request along to the handler
		handler(&loggableWriter, req, params)
		//Perform after-request hook
		serv.afterRequest(&loggableWriter, req)
	}
}

func (serv *Webserver) afterRequest(loggableWriter *LoggableResponseWriter, req *http.Request) {
	serv.routeHandler.logger.logLastRequest(req, loggableWriter.statusCode, loggableWriter.bytesWritten)
}

//Start starts the webserver
func (serv *Webserver) Start() {
	err := http.ListenAndServe(serv.listenAddr, serv.router)
	if err != nil {
		log.Fatal(err)
	}
}
