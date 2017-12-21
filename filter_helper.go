package qbin

import (
	"errors"
	"regexp"
	"strings"
)

//Disclaimer:
//Parts of the Filter code are copied from https://github.com/BlogSpam-Net/blogspam-api
//

//FilterSpam ->Filter content with different Filters to categories spam
func FilterSpam(doc *Document) error {
	Log.Warning("Spamfiltering is still in deveopement.")
	Log.Debug("Starting spamcheck.")

	//TODO: multiplier by linecount

	lines := strings.Count(doc.Content, "\n")
	//Count links with regex
	links := len(regexp.MustCompile("(https://|http://)?(www)?[a-z0-9]*\\.[a-z0-9]*/?").FindAllStringIndex(doc.Content, -1))

	//check if there are too many links:
	if links*10 > lines {
		Log.Error("Content was classified as Spam and will be safed to Spam Table")
		return errors.New("Internal Server Error")
	}
	return nil
}
