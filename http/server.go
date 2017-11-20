package qbinHTTP

import (
	"crypto/tls"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/qbin-io/backend"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/acme/autocert"
)

type Configuration struct {
	ListenHTTP    string
	ListenHTTPS   string
	FrontendPath  string
	Root          string
	Path          string
	CertWhitelist []string
	ForceRoot     bool
}

var config Configuration

// initializeConfig will normalize the options and create the "config" object.
func initializeConfig(initialConfig Configuration) {
	config = initialConfig
	// Transform frontendPath to an absolute path
	frontendPath, err := filepath.Abs(config.FrontendPath)
	if err != nil {
		qbin.Log.Critical("Frontend path couldn't be resolved.")
		panic(err)
	}
	config.FrontendPath = frontendPath

	config.Root = strings.TrimSuffix(config.Root, "/")

	// Extract "path" fron "root"
	rootParts := strings.SplitAfterN(config.Root, "/", 4) // https://example.org/[grab this part]
	config.Path = "/"
	if len(rootParts) > 3 { // Otherwise: application in root folder
		config.Path = "/" + rootParts[3]
	}
	config.Path = strings.TrimSuffix(config.Path, "/")
}

// StartHTTP launches the HTTP server which is responsible for the frontend and the HTTP API.
func StartHTTP(initialConfig Configuration) {
	// Configure
	qbin.Log.Debug("Initializing HTTP server...")
	initializeConfig(initialConfig)

	// Route
	qbin.Log.Debug("Setting up routes...")
	r := mux.NewRouter()
	setupRoutes(r)

	// Middlewares
	n := negroni.New(negroni.NewRecovery())
	// Add important headers
	n.UseHandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Server", "qbin")
	})
	// Redirect to root
	if config.ForceRoot {
		n.UseFunc(func(res http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
			if req.Host != config.domain || !strings.HasPrefix(req.URL.Path, config.path+"/") {
				if !strings.HasPrefix(req.URL.Path, config.path+"/") {
					res.Header().Add("Location", config.Root)
				} else {
					res.Header().Add("Location", config.Root+config.path+strings.TrimPrefix(req.URL.Path, config.path))
				}
			} else {
				next(res, req)
			}
		})
	}
	n.UseHandler(r)

	// Serve
	if config.ListenHTTPS != "none" {
		qbin.Log.Noticef("HTTPS server starting on %s for redirection", config.ListenHTTP)
		go listenHTTP(redirector{})
		if config.ListenHTTP != "none" {
			qbin.Log.Noticef("HTTPS server starting on %s, you should be able to reach it at %s", config.ListenHTTPS, config.Root)
			go listenHTTPS(n)
		}
	} else {
		qbin.Log.Noticef("HTTP server starting on %s, you should be able to reach it at %s", config.ListenHTTP, config.Root)
		go listenHTTP(n)
	}
}

func listenHTTPS(r http.Handler) {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.CertWhitelist...),
		Cache:      autocert.DirCache("certs"),
	}
	server := &http.Server{
		Addr:    config.ListenHTTPS,
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	err := server.ListenAndServeTLS("", "")
	if err != nil {
		qbin.Log.Criticalf("HTTPS server error: %s", err)
		panic(err)
	}
}

func listenHTTP(r http.Handler) {
	err := http.ListenAndServe(config.ListenHTTP, r)
	if err != nil {
		qbin.Log.Criticalf("HTTP server error: %s", err)
		panic(err)
	}
}

type redirector struct{}

func (redirector) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Server", "qbin")
	res.Header().Add("Location", config.Root+req.URL.EscapedPath())
	res.WriteHeader(301)
	res.Write([]byte(""))
}
