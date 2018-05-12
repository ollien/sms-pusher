package web

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

//LoggableResponseWriter allows access to otherwise hidden information within ResponseWriter
type LoggableResponseWriter struct {
	http.ResponseWriter
	statusCode     int
	bytesWritten   int
	headersWritten bool
}

//routeLogger is a logger that automatically includes route information.
type routeLogger struct {
	*logrus.Logger
	reasons map[*http.Request]string
}

//NewLoggableResponseWriter creats a LoggableResponseWriter with the given http.ResponseWriter
func NewLoggableResponseWriter(writer http.ResponseWriter) LoggableResponseWriter {
	return LoggableResponseWriter{
		ResponseWriter: writer,
	}
}

//Write is identical to http.ResponseWriter.Write, but stores the bytes sent and accounts for the 200 special case that Write normally handles by the interface's definition.
func (writer *LoggableResponseWriter) Write(bytes []byte) (int, error) {
	//Normally, this would be done by ResponseWriter.Write, but the wrapped response writer will not call this method, thus we have to force this same behavior.
	if !writer.headersWritten {
		writer.WriteHeader(http.StatusOK)
	}

	n, err := writer.ResponseWriter.Write(bytes)
	writer.bytesWritten += n

	return n, err
}

//WriteHeader is identical to http.Responsewriter.WriteHeader, but stores the status code.
func (writer *LoggableResponseWriter) WriteHeader(statusCode int) {
	if !writer.headersWritten {
		writer.headersWritten = true
		writer.statusCode = statusCode
		writer.ResponseWriter.WriteHeader(statusCode)
	}
}

func (logger *routeLogger) setResponseReason(req *http.Request, reason string) {
	logger.reasons[req] = reason
}

func (logger *routeLogger) setResponseErrorReason(req *http.Request, err error) {
	logger.setResponseReason(req, err.Error())
}

//logWithRoute returns a logrus.Entry that contains a field of the route that is being logged
func (logger *routeLogger) log(req *http.Request) *logrus.Entry {
	return logger.WithField(routeKey, req.RequestURI)
}

//logWithRouteField is equivalent to logrus.WithField, but inserts information about the route that is being logged.
func (logger *routeLogger) logWithField(req *http.Request, key string, value interface{}) *logrus.Entry {
	fields := make(logrus.Fields)
	fields[key] = value
	return logger.logWithFields(req, fields)
}

//logWithFields is equivalent to logrus.WithFields, but inserts information about the route that is being logged.
func (logger *routeLogger) logWithFields(req *http.Request, fields logrus.Fields) *logrus.Entry {
	fields[routeKey] = req.RequestURI
	return logger.WithFields(fields)
}

func (logger *routeLogger) logLastRequest(req *http.Request, statusCode int, bytesWritten int) {
	reason := logger.reasons[req]
	fields := logrus.Fields{
		"remote":      req.RemoteAddr,
		"proto":       req.Proto,
		"method":      req.Method,
		"user_agent":  req.UserAgent,
		"status_code": statusCode,
		"bytes":       bytesWritten,
	}
	logEntry := logger.WithFields(fields)
	if statusCode >= 400 && statusCode < 500 {
		logEntry.Warn(reason)
	} else if statusCode >= 500 {
		logEntry.Error(reason)
	} else {
		logEntry.Info(reason)
	}
}
