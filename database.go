package qbin

import (
	"database/sql"
	// MySQL/MariaDB Database Driver
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var isConnected bool

// Connect tries to establish a connection to a MySQL/MariaDB database under the given URI and initializes the qbin tables if they don't exist yet.
func Connect(uri string) error {
	Log.Noticef("Connecting to database at %s", uri)
	var err error
	db, err = sql.Open("mysql", uri)
	if err != nil {
		return err
	}

	// Print database version
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
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

	isConnected = true
	return nil
}

// IsConnected returns true if the database has already been initialized.
func IsConnected() bool {
	return isConnected
}
