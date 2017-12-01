package qbin

import (
	"testing"

	"github.com/qbin-io/backend"
)

func connect() {
	if !qbin.IsConnected() {
		err := qbin.Connect("root:@tcp(localhost)/qbin")
		if err != nil {
			panic(err)
		}
	}
}
func TestDocumentStorage(t *testing.T) {
	connect()

	exp, err := qbin.ParseExpiration("15m")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	doc := qbin.Document{
		Content:    "Hello <World>, &lt;this is a test&gt;",
		Syntax:     "",
		Expiration: exp,
		Address:    "::ffff:127.0.0.1",
	}

	err = qbin.Store(&doc)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("Stored document as %s", doc.ID)

	doc2, err := qbin.Request(doc.ID, true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if doc2.Content != doc.Content {
		t.Errorf("Content mismatch, received: %s (expected: %s)", doc2.Content, doc.Content)
	}
	if doc2.Address != doc.Address {
		t.Errorf("Address mismatch, received: %s (expected: %s)", doc2.Address, doc.Address)
	}
	if doc2.Syntax != doc.Syntax {
		t.Errorf("Syntax mismatch, received: %s (expected: %s)", doc2.Syntax, doc.Syntax)
	}
	if doc2.Upload != doc.Upload {
		t.Errorf("Upload mismatch, received: %s (expected %s)", doc2.Upload, doc.Upload)
	}
	if doc2.Expiration != doc.Expiration {
		t.Errorf("Expiration mismatch, received: %s (expected %s)", doc2.Expiration, doc.Expiration)
	}
}
