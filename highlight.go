package qbin

import (
	"io/ioutil"
	"net"
	"strings"
	"time"
)

var prismServer = "/tmp/prismjs.sock"
var languages = getLanguages()

// Highlight performs syntax highlighting on a string using Prism.js running under node.js.
func Highlight(content string, language string) (string, error) {
	var conn net.Conn
	var err error
	if strings.Contains(prismServer, "/") {
		conn, err = net.Dial("unix", prismServer)
	} else {
		conn, err = net.Dial("tcp", prismServer)
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

	ln := `<span class="line-number"></span>`
	return ln + strings.Replace(string(result), "\n", "\n"+ln, -1), nil
}

// SyntaxExists checks if a given syntax definition exists in Prism.js.
func SyntaxExists(language string) bool {
	return languages[language]
}

// ParseSyntax applies aliases and some other transformations to a syntax name supplied by the user to make it more intuitive.
func ParseSyntax(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	switch language {
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
func getLanguages() map[string]bool {
	result := map[string]bool{"": true}
	_, err := try(func() (interface{}, error) {
		// Get list of existing languages from prism-server
		languages, err := Highlight("", "list")
		if err != nil {
			return nil, err
		}
		list := strings.Split(languages, ",")

		// Set every existing language to true, the default value is false.
		for _, lang := range list {
			result[lang] = true
		}
		return result, nil
	}, 50, 100*time.Millisecond)
	if err != nil {
		Log.Errorf("Prism.js initialization failed: giving up on the following error: %s", err)
		return result
	}
	return result
}
