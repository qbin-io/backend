package qbin

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

const MaxFilesize = 1024 * 1024 // 1MB

// Document specifies the content and metadata of a piece of code that is hosted on qbin.
type Document struct {
	// ID is set on Store()
	ID      string
	Content string
	Syntax  string
	// Upload is set on Store()
	Upload     time.Time
	Expiration time.Time
	Address    string
	Views      int
}

// Store a document object in the database.
func Store(document *Document) error {
	// Generate a name that doesn't exist yet
	name, err := GenerateSafeName()
	if err != nil {
		return err
	}
	document.ID = name

	// Round the timestamps on the object. Won't affect the database, but we want consistency.
	document.Upload = time.Now().Round(time.Second)
	document.Expiration = document.Expiration.Round(time.Second)

	// Normalize new lines
	document.Content = strings.Trim(strings.Replace(strings.Replace(document.Content, "\r\n", "\n", -1), "\r", "\n", -1), "\n") + "\n"

	// Filter Content for Spam
	errS := FilterSpam(document)
	if errS != nil {
		return errS
	}

	// Don't accept binary files
	if strings.Contains(document.Content, "\x00") {
		return errors.New("file contains 0x00 bytes")
	}

	var content string
	// Consistency is power!
	if document.Syntax == "none" {
		document.Syntax = ""
	}
	content, err = Highlight(document.Content, document.Syntax)
	if err != nil {
		Log.Warningf("Skipped syntax highlighting for the following reason: %s", err)
	}

	var expiration interface{}
	if (document.Expiration != time.Time{}) {
		expiration = document.Expiration.UTC().Format("2006-01-02 15:04:05")
	}

	// Write the document to the database
	_, err = db.Exec(
		"INSERT INTO documents (id, content, syntax, upload, expiration, address, views) VALUES (?, ?, ?, ?, ?, ?, ?)",
		document.ID,
		content,
		document.Syntax,
		document.Upload.UTC().Format("2006-01-02 15:04:05"),
		expiration,
		document.Address,
		document.Views)
	if err != nil {
		return err
	}
	return nil
}

// Request a document from the database by its ID.
func Request(id string, raw bool) (Document, error) {
	doc := Document{ID: id}
	var views int
	var upload, expiration sql.NullString
	err := db.QueryRow("SELECT content, syntax, upload, expiration, address, views FROM documents WHERE id = ?", id).
		Scan(&doc.Content, &doc.Syntax, &upload, &expiration, &doc.Address, &views)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			Log.Warningf("Error retrieving document: %s", err)
		}
		return Document{}, err
	}

	go db.Exec("UPDATE documents SET views = views + 1 WHERE id = ?", id)
	doc.Views = views

	doc.Upload, _ = time.Parse("2006-01-02 15:04:05", upload.String)

	if expiration.Valid {
		doc.Expiration, err = time.Parse("2006-01-02 15:04:05", expiration.String)
		if doc.Expiration.Before(time.Unix(0, 1)) {
			if doc.Views > 0 {
				// Volatile document
				_, err = db.Exec("DELETE FROM documents WHERE id = ?", id)
				if err != nil {
					Log.Errorf("Couldn't delete volatile document: %s", err)
				}
			}
		} else {
			if err != nil {
				return Document{}, err
			}
			if doc.Expiration.Before(time.Now()) {
				return Document{}, errors.New("the document has expired")
			}
		}
	}

	if raw {
		doc.Content = StripHTML(doc.Content)
	}
	return doc, nil
}
