package db

import (
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	//DuplicateUserError is a postgres specific error for duplicate users in our users db
	DuplicateUserError = "pq: duplicate key value violates unique constraint \"users_username_key\""
	passwordCost       = 10
)

//CreateUser insersts a user into the database
func (db DatabaseConnection) CreateUser(username string, password []byte) error {
	hash, err := bcrypt.GenerateFromPassword(password, passwordCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users VALUES(DEFAULT, $1, $2)", username, hash)

	return err
}

//GetUser gets a user from the database and returns a User.
func (db DatabaseConnection) GetUser(username string) (User, error) {
	userRow := db.QueryRow("SELECT * FROM users WHERE username = $1", username)
	var id int
	var internalUsername string
	var passwordHash []byte
	err := userRow.Scan(&id, &internalUsername, &passwordHash)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:           id,
		Username:     internalUsername,
		passwordHash: passwordHash,
	}

	return user, nil
}

//GetUserByID gets a user from the database and returns a User.
func (db DatabaseConnection) GetUserByID(id int) (User, error) {
	userRow := db.QueryRow("SELECT * FROM users WHERE id = $1", id)
	var internalID int
	var username string
	var passwordHash []byte
	err := userRow.Scan(&internalID, &username, &passwordHash)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:           internalID,
		Username:     username,
		passwordHash: passwordHash,
	}

	return user, nil
}

//VerifyUser verifies a user against its authentication details. Returns the user if authed.
func (db DatabaseConnection) VerifyUser(username string, password []byte) (User, error) {
	user, err := db.GetUser(username)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword(user.passwordHash, password)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

//CreateSession makes a session given a User
func (db DatabaseConnection) CreateSession(user User) (Session, error) {
	sessionID := uuid.NewV4().String()
	_, err := db.Exec("INSERT INTO sessions VALUES($1, $2);", sessionID, user.ID)
	if err != nil {
		return Session{}, nil
	}

	return Session{
		ID:   sessionID,
		User: user,
	}, nil
}

//GetSession gets the user associated with a session
func (db DatabaseConnection) GetSession(sessionID string) (Session, error) {
	sessionRow := db.QueryRow("SELECT for_user FROM sessions WHERE id = $1", sessionID)
	var userID int
	err := sessionRow.Scan(&userID)
	if err != nil {
		return Session{}, err
	}

	user, err := db.GetUserByID(userID)
	if err != nil {
		return Session{}, err
	}

	return Session{
		ID:   sessionID,
		User: user,
	}, nil
}

//GetDevice gets a Device from the database, given a deviceID
func (db DatabaseConnection) GetDevice(deviceID uuid.UUID) (Device, error) {
	deviceRow := db.QueryRow("SELECT * FROM devices WHERE device_id = $1", deviceID)
	var id int
	var internalDeviceID uuid.UUID
	var fcmID []byte
	var userID int
	err := deviceRow.Scan(&id, &internalDeviceID, &fcmID, &userID)
	if err != nil {
		return Device{}, err
	}

	user, err := db.GetUserByID(userID)
	//If there's an error, there's an invalid user for the device. (i.e. doesn't exist)
	if err != nil {
		return Device{}, err
	}

	return Device{
		ID:       id,
		DeviceID: deviceID,
		FCMID:    fcmID,
		User:     user,
	}, nil

}

//RecordFile stores an MMS file to the database
func (db DatabaseConnection) RecordFile(fileName string, blockID uuid.UUID) error {
	_, err := db.Exec("INSERT INTO mms_files VALUES(DEFAULT, $1, $2);", fileName, blockID)

	return err
}

//RegisterDeviceToUser registers a device for a user
func (db DatabaseConnection) RegisterDeviceToUser(user User) (Device, error) {
	deviceID := uuid.NewV4()
	deviceRow := db.QueryRow("INSERT INTO devices VALUES(DEFAULT, $1, NULL, $2) RETURNING *;", deviceID, user.ID)
	var id int
	var internalDeviceID uuid.UUID
	var fcmID []byte
	var userID int
	err := deviceRow.Scan(&id, &internalDeviceID, &fcmID, &userID)
	if err != nil {
		return Device{}, err
	}

	return Device{
		ID:       id,
		DeviceID: deviceID,
		FCMID:    fcmID,
		User:     user,
	}, nil
}

//RegisterFCMID sets the FCM id (firebase_id) for a user's device, given a device id
func (db DatabaseConnection) RegisterFCMID(deviceID uuid.UUID, fcmID []byte) error {
	_, err := db.Exec("UPDATE devices SET firebase_id = $1 WHERE device_id = $2;", fcmID, deviceID)
	return err
}
