package qbin

import (
	"io/ioutil"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
)

// PrismServer defines the TCP address or unix socket path (containing a /) to prism-server.
var PrismServer = "/tmp/prism-server.sock"
var languages map[string]bool

// Highlight performs syntax highlighting on a string using Prism.js running under node.js.
func Highlight(content string, language string) (string, bool, error) {
	if language == "markdown!" {
		unsafe := blackfriday.Run([]byte(content))
		content = string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
		return content, true, nil
	}

	var conn net.Conn
	var err error
	if strings.Contains(PrismServer, "/") {
		conn, err = net.Dial("unix", PrismServer)
	} else {
		conn, err = net.Dial("tcp", PrismServer)
	}

	if err != nil {
		return content, false, err
	}

	_, err = conn.Write([]byte(language + "\n" + content))
	if err == nil {
		_, err = conn.Write([]byte{0})
	}
	if err != nil {
		return content, false, err
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil {
		return content, false, err
	}

	conn.Close()

	ln := ""
	if language != "list" {
		ln = `<span class="line-number"></span>`
	}
	// Don't ask where that 0x00 byte is coming from.
	// They are following me, and my code is haunted by them.
	// I guess it's just the closing character from the transmission though.
	return ln + strings.Replace(strings.TrimSuffix(string(result), "\x00"), "\n", "\n"+ln, -1), false, nil
}

// SyntaxExists checks if a given syntax definition exists in Prism.js.
func SyntaxExists(language string) bool {
	if language == "markdown!" {
		return true
	}
	if languages == nil {
		// Try getting the language list for the next document, seems like something broke or we're starting without prism-server.
		go getLanguages()
		return language == ""
	}
	return language == "" || languages[language]
}

// ParseSyntax applies aliases and some other transformations to a syntax name supplied by the user to make it more intuitive.
func ParseSyntax(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	switch language {
	case "none":
		return ""
	case "apache":
		return "apacheconf"
	case "c++":
		return "cpp"
	case "dockerfile":
		return "docker"
	case "html":
	case "htm":
	case "xml":
	case "svg":
		return "markup"
	case "js":
		return "javascript"
	case "golang":
		return "go"
	}
	return language
}

// getLanguages reads the existing languages from the prism-server for use with SyntaxExists.
var gettingLanguages = false

func getLanguages() {
	if gettingLanguages {
		return
	}
	gettingLanguages = true
	list := make([]string, 0)

	result, err := try(func() (interface{}, error) {
		// Get list of existing languages from prism-server
		result, _, err := Highlight("", "list")
		if err != nil {
			return nil, err
		}
		list = strings.Split(result, ",")

		// Set every existing language to true, the default value is false.
		return Slice2map(list), nil
	}, 120, 250*time.Millisecond)
	if err != nil {
		Log.Errorf("Prism.js initialization failed - giving up on the following error: %s", err)
		gettingLanguages = false
		return
	}
	sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
	Log.Debugf("Prism.js initialization succeeded. Available languages: %s", strings.Trim(strings.Replace(strings.Join(list, ", "), ", ,", ",", -1), ","))
	languages = result.(map[string]bool)
	gettingLanguages = false
}
