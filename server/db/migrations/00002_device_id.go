package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00002, Down00002)
}

func Up00002(tx *sql.Tx) error {
	_, err := tx.Exec("ALTER TABLE devices DROP COLUMN id;")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("ALTER TABLE devices RENAME COLUMN device_id TO id;")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("ALTER TABLE devices ADD PRIMARY KEY (id);")
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func Down00002(tx *sql.Tx) error {
	_, err := tx.Exec("ALTER TABLE devices DROP CONSTRAINT devices_pkey;")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("ALTER TABLE devices RENAME id TO device_id;")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("ALTER TABLE devices ADD COLUMN id SERIAL PRIMARY KEY;")
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
