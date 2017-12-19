package qbin

import (
	"errors"
	"strings"
)

//Disclaimer:
//Parts of the Filter code are copied from https://github.com/BlogSpam-Net/blogspam-api
//

//FilterSpam ->Filter content with different Filters to categories spam
func FilterSpam(doc *Document) error {
	Log.Warning("Spamfiltering is still in deveopement.")
	Log.Debug("Starting spamcheck.")
	lines := strings.Count(doc.Content, "\n")
	links := 0
	//check for "http://"
	links += strings.Count(doc.Content, "http://")
	//check for "https://"
	links += strings.Count(doc.Content, "https://")

	//check if there are too many links:
	if links*10 >= lines {
		Log.Error("Content was classified as Spam and will be safed to Spam Table")
		return errors.New("Internal Server Error")
	}
	return nil
}
