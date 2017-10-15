package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//RouteHandler holds all routes and allows them to share common variables
type RouteHandler struct {
	databaseConnection *sql.DB
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
	err := CreateUser(handler.databaseConnection, username, encodedPassword)

	if err != nil {
		//Postgres specific check
		if err.Error() == duplicateUserError {
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
	user, err = VerifyUser(handler.databaseConnection, username, encodedPassword)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessionID, err := CreateSession(handler.databaseConnection, user)
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

	deviceID, err := RegisterDeviceToUser(handler.databaseConnection, user)
	if err != nil {
		//TODO: Log data about 500
		writer.WriteHeader(http.StatusInternalServerError)
	}

	resultMap := make(map[string]string)
	resultMap["device_id"] = string(deviceID.DeviceID)
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
