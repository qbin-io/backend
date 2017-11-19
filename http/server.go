package qbinHTTP

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/qbin-io/backend"
)

type configuration struct {
	listen       string
	frontendPath string
	root         string
	path         string
}

var config configuration

// initializeConfig will normalize the options and create the "config" object.
func initializeConfig(listen string, frontendPath string, root string) {
	// Transform frontendPath to an absolute path
	frontendPath, err := filepath.Abs(frontendPath)
	if err != nil {
		qbin.Log.Critical("Frontend path couldn't be resolved.")
		panic(err)
	}

	root = strings.TrimSuffix(root, "/")

	// Extract "path" fron "root"
	rootParts := strings.SplitAfterN(root, "/", 4) // https://example.org/[grab this part]
	path := "/"
	if len(rootParts) > 3 { // Otherwise: application in root folder
		path = "/" + rootParts[3]
	}
	path = strings.TrimSuffix(path, "/")

	config = configuration{listen, frontendPath, root, path}
}

// StartHTTP launches the HTTP server which is responsible for the frontend and the HTTP API.
func StartHTTP(listen string, frontendPath string, root string) {
	// Configure
	qbin.Log.Debug("Initializing HTTP server...")
	initializeConfig(listen, frontendPath, root)

	// Route
	qbin.Log.Debug("Setting up routes...")
	r := mux.NewRouter()
	setupRoutes(r)

	// Serve
	qbin.Log.Noticef("HTTP server starting on %s, you should be able to reach it at %s", config.listen, config.root)
	err := http.ListenAndServe(config.listen, r)
	if err != nil {
		qbin.Log.Criticalf("HTTP server error: %s", err)
		panic(err)
	}
}
