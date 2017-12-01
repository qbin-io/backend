package qbin

import (
	"time"
)

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
}

// Store a document object in the database.
func Store(document *Document) error {
	// Generate a name that doesn't exist yet
	name, err := GenerateSafeName()
	if err != nil {
		return err
	}
	document.ID = name

	document.Upload = time.Now().Round(time.Second)
	document.Expiration = document.Expiration.Round(time.Second)

	var content string
	if document.Syntax == "none" {
		document.Syntax = ""
	}
	if document.Syntax != "" {
		// TODO: Highlight
	} else {
		content = escapeHTML(document.Content)
	}

	// Write the document to the database
	_, err = db.Exec(
		"INSERT INTO documents (id, content, syntax, upload, expiration, address) VALUES (?, ?, ?, FROM_UNIXTIME(?), FROM_UNIXTIME(?), ?)",
		document.ID,
		content,
		document.Syntax,
		document.Upload.Unix(),
		document.Expiration.Unix(),
		document.Address)
	if err != nil {
		return err
	}
	return nil
}

// Request a document from the database by its ID.
func Request(id string, raw bool) (Document, error) {
	doc := Document{ID: id}
	var upload, expiration int64
	err := db.QueryRow("SELECT content, syntax, UNIX_TIMESTAMP(upload), UNIX_TIMESTAMP(expiration), address FROM documents WHERE id = ?", id).
		Scan(&doc.Content, &doc.Syntax, &upload, &expiration, &doc.Address)
	if err != nil {
		return Document{}, err
	}

	doc.Upload = time.Unix(upload, 0)
	doc.Expiration = time.Unix(expiration, 0)

	if raw {
		doc.Content = stripHTML(doc.Content)
	}
	return doc, err
}
