package qbin

import (
	"errors"
	"regexp"
)

var blacklist = []regexp.Regexp{}

func spamcheckBlacklist(doc *Document) error {
	for i := 0; i < len(blacklist); i++ {
		if blacklist[i].MatchString(doc.Content) {
			return errors.New("document matches blacklist (#" + string(i) + ")")
		}
	}
	return nil
}

func spamcheckLinkCount(doc *Document) error {
	//Count word and determin, how many links are allowed in the document
	wordCount := len(regexp.MustCompile("\\w*\\S+\\w*").FindAllStringIndex(doc.Content, -1))
	var allowedLinks int
	switch {
	case wordCount > 500:
		allowedLinks = wordCount
	case wordCount > 123:
		allowedLinks = 30 + wordCount/10
	case wordCount > 50:
		allowedLinks = (wordCount / 8)
	case wordCount > 10:
		allowedLinks = 2 + (wordCount / 15)
	case wordCount == 1:
		allowedLinks = 0
	default:
		allowedLinks = 1
	}

	//Count links with regex
	links := len(regexp.MustCompile("(https://|http://)?(www)?([a-z0-9]+\\.)+[a-z0-9]+/?").FindAllStringIndex(doc.Content, -1))

	Log.Debugf("words: %i \n links: %i", wordCount, links)
	//check if there are too many links:
	if links > allowedLinks {
		return errors.New("")
	}
}

//FilterSpam ->Filter content with different Filters to categories spam
func FilterSpam(doc *Document) error {
	if err := spamcheckBlacklist(doc); err != nil {
		go saveToSpam(doc)
		return err
	}

	if err := spamcheckLinkCount(doc); err != nil {
		go saveToSpam(doc)
		return err
	}

	return nil
}

func saveToSpam(doc *Document) {
	_, err := db.Exec(
		"INSERT INTO spam (id, content, upload, address) VALUES (?, ?, ?, ?)",
		doc.ID,
		doc.Content,
		doc.Upload.UTC().Format("2006-01-02 15:04:05"),
		doc.Address)
	if err != nil {
		Log.Errorf("An error occured while saving spam to spam-DB: %s", err)
	}
	Log.Debug("Spam was saved to DB.")
}
