package web

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ollien/sms-pusher/server/db"
	"github.com/ollien/sms-pusher/server/firebasexmpp"
	uuid "github.com/satori/go.uuid"
)

//RouteHandler holds all routes and allows them to share common variables
type RouteHandler struct {
	databaseConnection *sql.DB
	sendChannel        chan<- firebasexmpp.OutboundMessage
	//TODO: add sendErrorChannel once websockets are implemented
}

func (handler RouteHandler) index(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	io.WriteString(writer, "<h1>Hello world!</h1>")
}

func (handler RouteHandler) register(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	username := req.FormValue("username")
	password := req.FormValue("password")

	if username == "" || password == "" || len(password) < 8 {
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encodedPassword := []byte(password)
	err := db.CreateUser(handler.databaseConnection, username, encodedPassword)

	if err != nil {
		//Postgres specific check
		if err.Error() == db.DuplicateUserError {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
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
		//TODO: Return data explaining why a 400 was returned
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	encodedPassword := []byte(password)
	user, err = db.VerifyUser(handler.databaseConnection, username, encodedPassword)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessionID, err := db.CreateSession(handler.databaseConnection, user)
	if err != nil {
		//TODO: Log data about 500
		writer.WriteHeader(http.StatusInternalServerError)
	}

	resultMap := make(map[string]string)
	resultMap["session_id"] = sessionID
	resultJSON, err := json.Marshal(resultMap)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		_, err := writer.Write(resultJSON)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
		} else {
			cookie := &http.Cookie{
				Name:  "session",
				Value: sessionID,
			}
			http.SetCookie(writer, cookie)
		}
	}
}

func (handler RouteHandler) registerDevice(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	user, err := GetSessionUser(handler.databaseConnection, req)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	deviceID, err := db.RegisterDeviceToUser(handler.databaseConnection, user)
	if err != nil {
		//TODO: Log data about 500
		writer.WriteHeader(http.StatusInternalServerError)
	}

	resultMap := make(map[string]string)
	resultMap["device_id"] = deviceID.DeviceID.String()
	resultJSON, err := json.Marshal(resultMap)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		_, err := writer.Write(resultJSON)
		if err != nil {
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
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	deviceUUID, err := uuid.FromString(deviceID)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	//Check to make sure that the user is actually modifying their device
	device, err := db.GetDevice(handler.databaseConnection, deviceUUID)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if device.DeviceUser.ID != user.ID {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = db.RegisterFCMID(handler.databaseConnection, deviceUUID, []byte(fcmID))
	if err != nil {
		//TODO: Log why a 500 was returned.
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
