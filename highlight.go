package qbin

import (
	"io/ioutil"
	"net"
	"strings"
)

var prismServer = "/tmp/prismjs.sock"
var languages = getLanguages()

func Highlight(content string, language string) string {
	var conn net.Conn
	var err error
	if strings.Contains(prismServer, "/") {
		conn, err = net.Dial("unix", prismServer)
	} else {
		conn, err = net.Dial("tcp", prismServer)
	}

	if err != nil {
		Log.Errorf("Couldn't connect to prism.js server: %s", err)
		return content
	}

	_, err = conn.Write([]byte(language + "\n" + content))
	if err == nil {
		_, err = conn.Write([]byte{0})
	}
	if err != nil {
		Log.Errorf("Couldn't write to prism.js server: %s", err)
		return content
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil {
		Log.Errorf("Couldn't read from connection to prism.js server: %s", err)
		return content
	}

	conn.Close()

	return string(result)
}

func SyntaxExists(language string) bool {
	return languages[language]
}

func getLanguages() map[string]bool {
	m := map[string]bool{}
	list := strings.Split(Highlight("", "list"), ",")
	for _, lang := range list {
		m[lang] = true
	}
	return m
}
