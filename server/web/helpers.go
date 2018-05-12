package web

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/ollien/sms-pusher/server/db"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/h2non/filetype.v1"
	fttypes "gopkg.in/h2non/filetype.v1/types"
)

const (
	uploadedFileMode = 0644
	routeKey         = "_route"
)

//GetSessionCookie gets the cookie named "session" from http.Cookies()
func GetSessionCookie(req *http.Request) *http.Cookie {
	for _, cookie := range req.Cookies() {
		if cookie.Name == "session" {
			return cookie
		}
	}

	return nil
}

//GetSessionUser gets the user associated with a session within a *http.Request.
func GetSessionUser(databaseConnection db.DatabaseConnection, req *http.Request) (db.User, error) {
	cookie := GetSessionCookie(req)
	sessionID := req.FormValue("session_id")
	if sessionID == "" {
		if cookie != nil {
			sessionID = cookie.Value
		} else {
			return db.User{}, errors.New("no session cookie found")
		}
	}

	session, err := databaseConnection.GetSession(sessionID)

	//If err is not nil, there is no valid session.
	return session.User, err
}

//StoreFile stores an incoming file to disk, with its SHA256 as its username
func StoreFile(databaseConnection db.DatabaseConnection, uploadLocation string, blockID uuid.UUID, bytes []byte) (int, error) {
	fileName, err := getFileName(bytes)
	if err != nil {
		return 0, err
	}

	databaseConnection.RecordFile(fileName, blockID)

	filePath := path.Join(uploadLocation, fileName)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, uploadedFileMode)
	if err != nil {
		return 0, err
	}

	return file.Write(bytes)
}

//Get the name for a file that we will be storing
func getFileName(bytes []byte) (string, error) {
	fileHash := sha256.Sum256(bytes)
	theType, err := filetype.Match(bytes)
	if err != nil {
		return "", err
	}

	extension := theType.Extension
	//If we can't figure out the type, assume it's a text entry
	if theType.MIME == (fttypes.MIME{}) {
		extension = "txt"
	}

	return fmt.Sprintf("%x.%s", fileHash, extension), nil
}

//setStatusTo500IfDatabaseFault writes a 500 status code if the error is a database fault. Otherwise, it writes the given status code.
func setStatusTo500IfDatabaseFault(writer http.ResponseWriter, err error, alternateStatusCode int) {
	if dbErr, ok := err.(db.DatabaseError); ok && dbErr.DatabaseFault {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(alternateStatusCode)
	}

}
