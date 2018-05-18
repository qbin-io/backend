package qbin

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strings"
)

var FilterEnable = map[string]bool{}
var domainBlacklist = []*regexp.Regexp{}
var urlBlacklist = []*regexp.Regexp{}
var contentBlacklist = []*regexp.Regexp{}

func LoadBlacklistFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		source := strings.Split(string(content), "\n")
		domainBlacklist = make([]*regexp.Regexp, 0)
		urlBlacklist = make([]*regexp.Regexp, 0)
		contentBlacklist = make([]*regexp.Regexp, 0)
		n := 0
		for _, expression := range source {
			if len(expression) > 0 && !strings.HasPrefix(expression, "#") {
				n++
				if strings.HasPrefix(expression, "[D] ") {
					domainBlacklist = append(domainBlacklist, regexp.MustCompile(strings.TrimPrefix(expression, "[D] ")))
				} else if strings.HasPrefix(expression, "[U] ") {
					urlBlacklist = append(urlBlacklist, regexp.MustCompile(strings.TrimPrefix(expression, "[U] ")))
				} else {
					contentBlacklist = append(contentBlacklist, regexp.MustCompile(expression))
				}
			}
		}

		Log.Debugf("%d blacklist expressions loaded.", n)
	}
	return err
}

func spamcheckBlacklist(doc *Document, highlighted *string) error {
	for i := 0; i < len(contentBlacklist); i++ {
		if contentBlacklist[i].MatchString(doc.Content) {
			return errors.New("document matches blacklist")
		}
	}

	links := linkExpression.FindAllStringSubmatch(*highlighted, -1)
	for i := 0; i < len(links); i++ {
		for j := 0; j < len(urlBlacklist); j++ {
			if urlBlacklist[j].MatchString(links[i][1]) {
				return errors.New("document matches blacklist")
			}
		}
		domain := domainExpression.FindStringSubmatch(links[i][1])[1]
		for j := 0; j < len(domainBlacklist); j++ {
			if domainBlacklist[j].MatchString(domain) {
				return errors.New("document matches blacklist")
			}
		}
	}
	return nil
}

var spacesExpression = regexp.MustCompile(`\s+`)
var linkExpression = regexp.MustCompile(`url-link" href="([^"]+)"`)
var domainExpression = regexp.MustCompile(`^[^:]+://([^/]+)`)

func spamcheckLinkCount(doc *Document, highlighted *string) error {
	//Count word and determin, how many links are allowed in the document
	documentLength := len(spacesExpression.ReplaceAllString(doc.Content, ""))

	links := linkExpression.FindAllStringSubmatch(*highlighted, -1)
	linkCount := len(links)
	linkLength := 0
	for i := 0; i < linkCount; i++ {
		linkLength += len(links[i][1])
	}

	documentLength -= linkLength

	var allowedLinkCount = 3
	var allowedLinkRatio = 25

	if documentLength < 100 {
		// Under 100 characters: only 2 links allowed
		allowedLinkCount = 2
		allowedLinkRatio = 0
	} else if documentLength < 400 {
		// always allow up to 2 links, more if it's not more than 45% of the document
		allowedLinkCount = 2
		allowedLinkRatio = 45
	} else if documentLength < 500 {
		// always allow up to 2 links, otherwise linear approximation
		allowedLinkCount = 2
		allowedLinkRatio = 25 + (500-documentLength)*20/100
	}

	if linkCount > allowedLinkCount && linkLength*100/(documentLength+linkLength) > allowedLinkRatio {
		return errors.New("document has too many links")
	}
	return nil
}

//FilterSpam ->Filter content with different Filters to categories spam
func FilterSpam(doc *Document, highlighted *string) error {

	if FilterEnable["blacklist"] {
		err := spamcheckBlacklist(doc, highlighted)
		if err != nil {
			go saveToSpam(doc)
			return err
		}
	}

	if FilterEnable["linkcount"] {
		err := spamcheckLinkCount(doc, highlighted)
		if err != nil {
			go saveToSpam(doc)
			return err
		}
	}

	return nil
}

func saveToSpam(doc *Document) {
	_, err := db.Exec(
		"INSERT INTO spam (id, content, upload) VALUES (?, ?, ?)",
		doc.ID,
		doc.Content,
		doc.Upload.UTC().Format("2006-01-02 15:04:05"))
	if err != nil {
		Log.Errorf("An error occured while saving spam to spam-DB: %s", err)
	}
	Log.Debug("Spam was saved to DB.")
}
