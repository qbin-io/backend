package qbin

import (
	"io/ioutil"
	"net"
	"sort"
	"strings"
	"time"
)

// PrismServer defines the TCP address or unix socket path (containing a /) to prism-server.
var PrismServer = "/tmp/prism-server.sock"
var languages map[string]bool

// Highlight performs syntax highlighting on a string using Prism.js running under node.js.
func Highlight(content string, language string) (string, error) {
	var conn net.Conn
	var err error
	if strings.Contains(PrismServer, "/") {
		conn, err = net.Dial("unix", PrismServer)
	} else {
		conn, err = net.Dial("tcp", PrismServer)
	}

	if err != nil {
		return content, err
	}

	_, err = conn.Write([]byte(language + "\n" + content))
	if err == nil {
		_, err = conn.Write([]byte{0})
	}
	if err != nil {
		return content, err
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil {
		return content, err
	}

	conn.Close()

	ln := ""
	if language != "list" {
		ln = `<span class="line-number"></span>`
	}
	// Don't ask where that 0x00 byte is coming from.
	// They are following me, and my code is haunted by them.
	// I guess it's just the closing character from the transmission though.
	return ln + strings.Replace(strings.TrimSuffix(string(result), "\x00"), "\n", "\n"+ln, -1), nil
}

// SyntaxExists checks if a given syntax definition exists in Prism.js.
func SyntaxExists(language string) bool {
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
	languageSlice := make([]string, 0)

	result, err := try(func() (interface{}, error) {
		// Get list of existing languages from prism-server
		result, err := Highlight("", "list")
		if err != nil {
			return nil, err
		}
		list := strings.Split(result, ",")

		// Set every existing language to true, the default value is false.
		languages := make(map[string]bool)
		for _, lang := range list {
			languages[lang] = true
			if lang != "" {
				languageSlice = append(languageSlice, lang)
			}
		}
		return languages, nil
	}, 120, 250*time.Millisecond)
	if err != nil {
		Log.Errorf("Prism.js initialization failed - giving up on the following error: %s", err)
		gettingLanguages = false
		return
	}
	sort.Slice(languageSlice, func(i, j int) bool { return languageSlice[i] < languageSlice[j] })
	Log.Debugf("Prism.js initialization succeeded. Available languages: %s", strings.Join(languageSlice, ", "))
	languages = result.(map[string]bool)
	gettingLanguages = false
}
