package qbin

import (
	"io/ioutil"
	"net"
	"strings"
	"time"
)

var prismServer = "/tmp/prismjs.sock"
var languages = getLanguages()

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

	return string(result), nil
}

func SyntaxExists(language string) bool {
	return languages[language]
}

// getLanguages reads the existing languages from the prism-server for use with SyntaxExists.
func getLanguages() map[string]bool {
	result, err := try(func() (interface{}, error) {
		// Get list of existing languages from prism-server
		languages, err := Highlight("", "list")
		if err != nil {
			return nil, err
		}
		list := strings.Split(languages, ",")

		// Set every existing language to true, the default value is false.
		m := map[string]bool{}
		for _, lang := range list {
			m[lang] = true
		}
		return m, nil
	}, 50, 100*time.Millisecond)
	if err != nil {
		Log.Errorf("Prism.js initialization failed: giving up on the following error: %s", err)
		return map[string]bool{}
	}
	return result.(map[string]bool)
}
