package qbin

import (
	"database/sql"
	"time"
	// MySQL/MariaDB Database Driver
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var isConnected bool

// Connect tries to establish a connection to a MySQL/MariaDB database under the given URI and initializes the qbin tables if they don't exist yet.
func Connect(uri string) error {
	Log.Noticef("Connecting to database at %s", uri)
	result, err := try(func() (interface{}, error) {
		return sql.Open("mysql", uri)
	}, 10, time.Second) // Wait up to 10 seconds for the database
	if err != nil {
		return err
	}
	db = result.(*sql.DB)

	// Print database version
	var version string
	_, err = try(func() (interface{}, error) {
		return nil, db.QueryRow("SELECT VERSION()").Scan(&version)
	}, 10, time.Second) // Wait up to 10 seconds for the database
	if err != nil {
		return err
	}
	Log.Noticef("Database version: %s", version)

	// Create tables
	var table string
	db.QueryRow("SHOW TABLES LIKE 'documents'").Scan(&table)
	if table == "" {
		Log.Noticef("Setting up database...")
		err = db.QueryRow(`CREATE TABLE documents (
            id varchar(30) PRIMARY KEY,
            content longtext NOT NULL,
            syntax varchar(30) NOT NULL DEFAULT "",
            upload timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
            expiration timestamp NULL,
            address varchar(45) NOT NULL
        )`).Scan()
		if err != nil && err.Error() != "sql: no rows in result set" {
			return err
		}
		Log.Noticef("Created table `documents`. Database setup completed.")
	}

	safeName, errSafeName = db.Prepare("SELECT COUNT(id) FROM documents WHERE id = ?")

	isConnected = true
	go cleanup()
	return nil
}

// IsConnected returns true if the database has already been initialized.
func IsConnected() bool {
	return isConnected
}

func cleanup() {
	stmt, err := db.Prepare("DELETE FROM documents WHERE expiration < CURRENT_TIMESTAMP")
	if err != nil {
		Log.Errorf("Couldn't initialize cleanup statement: %s", err)
		return
	}

	for {
		result, err := stmt.Exec()
		if err != nil {
			Log.Errorf("Couldn't execute cleanup statement: %s", err)
		} else {
			n, err := result.RowsAffected()
			if err == nil && n > 0 {
				Log.Debugf("Cleaned up %d documents.", n)
			}
		}

		time.Sleep(10 * time.Minute)
	}
}
