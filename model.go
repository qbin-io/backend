package qbin

import "time"

// Document specifies the content and metadata of a piece of code that is hosted on qbin.
type Document struct {
	id         string
	content    string
	syntax     string
	upload     time.Time
	expiration time.Time
	address    string
}

// Create a new document object from its source.
func Create(content string, syntax string, expiration string, address string) Document {
	return Document{}
}

// Store a document object in the database.
func Store(document Document) error {
	return nil
}

// Request a document from the database by its ID.
func Request(id string, raw bool) (Document, error) {
	return Document{}, nil
}
