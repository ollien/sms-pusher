package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ollien/sms-pusher/server/config"
	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	"github.com/sirupsen/logrus"
)

const (
	//4 KB
	maxRequestSize = 4096
	//16 MB
	maxFileSize = 16777216
)

//hookedRouter allows for us to direct http requests to the correct handlers with the addition of pre/post request hooks
//Performing these hooks within the hookedRouter rather than when wrapping the handler functions allows us to properly handle all status codes including 404.
type hookedRouter struct {
	BeforeRequest func(http.ResponseWriter, *http.Request)
	AfterRequest  func(http.ResponseWriter, *http.Request)
	httprouter.Router
}

//Webserver hosts a webserver for sms-pusher
type Webserver struct {
	logger       *logrus.Logger
	routeHandler RouteHandler
	Server       http.Server
}

type loggableHandlerFunction = func(*LoggableResponseWriter, *http.Request, httprouter.Params)

//handlerFunction is what httprouter is expecting as a handler function. Before being used, all loggableHandlerFunctions will be wrapped as a handlerFunction such that they are compatiable with httprouter.
type handlerFunction = func(http.ResponseWriter, *http.Request, httprouter.Params)

//NewWebserver creats a new Webserver with httpServer being set to a new http.Server
func NewWebserver(listenAddr string, databaseConnection db.DatabaseConnection, sendChannel chan<- firebasexmpp.DownstreamPayload, logger *logrus.Logger) (Webserver, error) {
	config, err := config.GetConfig()
	if err != nil {
		return Webserver{}, err
	}

	routeHandler := RouteHandler{
		databaseConnection: databaseConnection,
		sendChannel:        sendChannel,
		logger:             newRouteLogger(logger),
	}
	router := newRouter()
	httpServer := http.Server{
		Addr:    config.Web.GetListenAddress(),
		Handler: router,
	}
	webserver := Webserver{
		logger:       logger,
		routeHandler: routeHandler,
		Server:       httpServer,
	}
	router.AfterRequest = webserver.afterRequest
	webserver.initHandlers()

	return webserver, nil
}

func (serv *Webserver) initHandlers() {
	router := serv.Server.Handler.(*hookedRouter)
	router.GET("/", serv.wrapHandlerFunction(serv.routeHandler.index))
	router.POST("/register", serv.wrapHandlerFunction(serv.routeHandler.register))
	router.POST("/authenticate", serv.wrapHandlerFunction(serv.routeHandler.authenticate))
	router.POST("/register_device", serv.wrapHandlerFunction(serv.routeHandler.registerDevice))
	router.POST("/set_fcm_id", serv.wrapHandlerFunction(serv.routeHandler.setFCMID))
	router.POST("/send_message", serv.wrapHandlerFunction(serv.routeHandler.sendMessage))
	router.POST("/upload_mms_file", serv.wrapHandlerFunctionWithLimit(serv.routeHandler.uploadMMSFile, maxFileSize))
}

//wrapHandlerFunction allows us to enforce a file size limit
//Though we could theoretically put this in ServeHTTP, this allows us to set different sizes for different routes after httprouter has taken care of the route handling for us.
func (serv *Webserver) wrapHandlerFunction(handler loggableHandlerFunction) handlerFunction {
	return serv.wrapHandlerFunctionWithLimit(handler, maxRequestSize)
}

//wrapHandlerFunctionWithLimit is the same as wrapHandlerFunction but allows us to set a size limit on the request
func (serv *Webserver) wrapHandlerFunctionWithLimit(handler loggableHandlerFunction, sizeLimit int64) handlerFunction {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		//Due to our wrapping within hookedRouter.ServeHTTP, we can always expect a LoggableResponseWriter
		loggableWriter := writer.(*LoggableResponseWriter)
		//Enforce a max file size
		req.Body = http.MaxBytesReader(loggableWriter, req.Body, sizeLimit)
		//Pass our request to the handler only if we have a valid form.
		if serv.checkFormValidity(loggableWriter, req) {
			handler(loggableWriter, req, params)
		}
	}
}

//checkFormValidity checks if a given form is valid. If it's not, 403 is written and false is returned. If it is, true is retruned and the header remains unchanged.
func (serv *Webserver) checkFormValidity(writer *LoggableResponseWriter, req *http.Request) bool {
	err := req.ParseForm()
	if err != nil {
		writer.setResponseErrorReason(err)
		writer.WriteHeader(http.StatusBadRequest)
		return false
	}
	return true
}

func (serv *Webserver) afterRequest(writer http.ResponseWriter, req *http.Request) {
	if loggableWriter, ok := writer.(*LoggableResponseWriter); ok {
		serv.routeHandler.logger.logLastRequest(req, loggableWriter.statusCode, loggableWriter.responseReason, loggableWriter.bytesWritten)
	}
}

//NewRouter creates a new Router[:w
func newRouter() *hookedRouter {
	return &hookedRouter{
		Router: *httprouter.New(),
	}
}

func (router *hookedRouter) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	loggableWriter := NewLoggableResponseWriter(writer)
	if router.BeforeRequest != nil {
		router.BeforeRequest(&loggableWriter, req)
	}
	router.Router.ServeHTTP(&loggableWriter, req)
	if router.AfterRequest != nil {
		router.AfterRequest(&loggableWriter, req)
	}
}
