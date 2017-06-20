package main

import "encoding/json"
import "database/sql"
import _ "github.com/lib/pq"
import "os"

const CONFIG_URI_KEY = "uri"

func initDb(configPath string) (*sql.DB, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	jsonDecoder := json.NewDecoder(file)
	configMap := make(map[string]string)
	jsonDecoder.Decode(&configMap)
	db, err := sql.Open("postgres", configMap[CONFIG_URI_KEY])
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (" +
			"id SERIAL," +
			"phone_number VARCHAR(16)," +
			"time timestamp," +
			"message TEXT);")
	if err != nil {
		return nil, err
	}
	return db, nil
}
