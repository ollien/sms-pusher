package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00001, Down00001)
}

func Up00001(tx *sql.Tx) error {
	//Create users table.
	_, err := tx.Exec("CREATE TABLE users (" +
		"id SERIAL PRIMARY KEY," +
		"username VARCHAR(32) UNIQUE," +
		"password_hash bytea);")
	if err != nil {
		return err
	}

	//Create devices table
	_, err = tx.Exec("CREATE TABLE devices (" +
		"id SERIAL PRIMARY KEY," +
		"device_id uuid UNIQUE," +
		"firebase_id bytea," +
		"for_user INTEGER REFERENCES users(id));")
	if err != nil {
		return err
	}

	//Create sessions table
	_, err = tx.Exec("CREATE TABLE sessions (" +
		"id uuid PRIMARY KEY," +
		"for_user INTEGER REFERENCES users(id));")
	if err != nil {
		return err
	}

	//Create mms_file_blocks table
	_, err = tx.Exec("CREATE TABLE mms_file_blocks(" +
		"id uuid PRIMARY KEY," +
		"for_user INTEGER REFERENCES users(id));")
	if err != nil {
		return err
	}

	//Create mms_files table
	//name is 128 bytes, as the file hash will be 64 bytes, plus we need room for the extension. Rounded up to the nearest power of two, 128.
	_, err = tx.Exec("CREATE TABLE mms_files(" +
		"id SERIAL PRIMARY KEY," +
		"name VARCHAR(128)," +
		"block uuid REFERENCES mms_file_blocks(id));")
	if err != nil {
		return err
	}

	return nil
}

func Down00001(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE devices;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE sessions;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE mms_files;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE mms_file_blocks;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE users;")
	if err != nil {
		return err
	}

	return nil
}
