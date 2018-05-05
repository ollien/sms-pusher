package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/ollien/sms-pusher/server/config"
	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	databaseErrorLogFormat   = "Database error: %s"
	jsonErrorLogFormat       = "JSON error: %s"
	xmppErrorLogFormat       = "XMPP error: %s"
	badUUIDErrorLogMsg       = "Bad UUID for device"
	notEnoughInfoErrorLogMsg = "Not enough info to continue."
	badB64ErrorLogMsg        = "Invalid B64 for uploaded file."
)

//RouteHandler holds all routes and allows them to share common variables
type RouteHandler struct {
	databaseConnection db.DatabaseConnection
	sendChannel        chan<- firebasexmpp.OutboundMessage
	logger             *logrus.Logger
	//TODO: We really don't need to have an entire field for this, and should probably refactor config to be a bit cleaner about things like this...
	mmsConfig config.MMSConfig
	//TODO: add sendErrorChannel once websockets are implemented
}

func (handler RouteHandler) index(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}

func (handler RouteHandler) register(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	username := req.FormValue("username")
	password := req.FormValue("password")
	if username == "" || password == "" || len(password) < 8 {
		logWithRoute(handler.logger, req).Debug(notEnoughInfoErrorLogMsg)
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encodedPassword := []byte(password)
	err := handler.databaseConnection.CreateUser(username, encodedPassword)
	if err != nil {
		//Postgres specific check
		if err.Error() == db.DuplicateUserError {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
			logWithRoute(handler.logger, req).Errorf(databaseErrorLogFormat, err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (handler RouteHandler) authenticate(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := GetSessionUser(handler.databaseConnection, req)
	//If there is no error, we found a user, and can return a 200
	if err == nil {
		//If there is a valid session, we have a 200, which is already the default header, so we just reurn.
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")
	if username == "" || password == "" {
		logWithRoute(handler.logger, req).Debug(notEnoughInfoErrorLogMsg)
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encodedPassword := []byte(password)
	user, err = handler.databaseConnection.VerifyUser(username, encodedPassword)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, err := handler.databaseConnection.CreateSession(user)
	if err != nil {
		logWithRoute(handler.logger, req).Errorf(databaseErrorLogFormat, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:  "session",
		Value: session.ID,
	}

	http.SetCookie(writer, cookie)
}

func (handler RouteHandler) registerDevice(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := GetSessionUser(handler.databaseConnection, req)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	deviceID, err := handler.databaseConnection.RegisterDeviceToUser(user)
	if err != nil {
		logWithRoute(handler.logger, req).Errorf(databaseErrorLogFormat, err)
		writer.WriteHeader(http.StatusInternalServerError)
	}

	resultMap := make(map[string]string)
	resultMap["device_id"] = deviceID.DeviceID.String()
	resultJSON, err := json.Marshal(resultMap)
	if err != nil {
		resultMapString := fmt.Sprintf("%#v", resultMap)
		logWithRouteField(handler.logger, req, "map", resultMapString).Errorf(jsonErrorLogFormat, err)
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		_, err := writer.Write(resultJSON)
		if err != nil {
			logWithRouteField(handler.logger, req, "json", string(resultJSON)).Errorf(jsonErrorLogFormat, err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (handler RouteHandler) setFCMID(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := GetSessionUser(handler.databaseConnection, req)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	deviceID := req.FormValue("device_id")
	fcmID := req.FormValue("fcm_id")
	if deviceID == "" || fcmID == "" {
		logWithRoute(handler.logger, req).Debug(notEnoughInfoErrorLogMsg)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	deviceUUID, err := uuid.FromString(deviceID)
	if err != nil {
		logWithRoute(handler.logger, req).Debug(badUUIDErrorLogMsg)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	//Check to make sure that the user is actually modifying their device
	device, err := handler.databaseConnection.GetDevice(deviceUUID)
	if err != nil {
		logWithRoute(handler.logger, req).Errorf(databaseErrorLogFormat, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if device.User.ID != user.ID {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = handler.databaseConnection.RegisterFCMID(deviceUUID, []byte(fcmID))
	if err != nil {
		logWithRoute(handler.logger, req).Errorf(databaseErrorLogFormat, err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func (handler RouteHandler) sendMessage(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := GetSessionUser(handler.databaseConnection, req)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	recipient := req.FormValue("recipient")
	message := req.FormValue("message")
	deviceID := req.FormValue("device-id")
	if recipient == "" || message == "" || deviceID == "" {
		logWithRoute(handler.logger, req).Debug(notEnoughInfoErrorLogMsg)
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	deviceUUID, err := uuid.FromString(deviceID)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	device, err := handler.databaseConnection.GetDevice(deviceUUID)
	if device.User.ID != user.ID {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	smsMessage := firebasexmpp.SMSMessage{
		PhoneNumber: recipient,
		Message:     message,
		Timestamp:   time.Now().Unix(),
	}
	outMessage, err := firebasexmpp.ConstructDownstreamSMS(device.FCMID, smsMessage)
	if err != nil {
		logWithRoute(handler.logger, req).Errorf(xmppErrorLogFormat, err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	handler.sendChannel <- outMessage
}

func (handler RouteHandler) uploadMMSFile(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := GetSessionUser(handler.databaseConnection, req)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	//TODO: check if supported MIME type

	b64 := req.FormValue("data")
	deviceID := req.FormValue("device-id")
	if b64 == "" || deviceID == "" {
		logWithRoute(handler.logger, req).Debug(notEnoughInfoErrorLogMsg)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	b64Encoding := base64.NewEncoding("base64")
	fileBytes, err := b64Encoding.DecodeString(b64)
	if err != nil {
		logWithRoute(handler.logger, req).Error(badB64ErrorLogMsg)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = StoreFile(handler.databaseConnection, handler.mmsConfig.UploadLocation, user, fileBytes)
	if err != nil {
		logWithRoute(handler.logger, req).Error(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
