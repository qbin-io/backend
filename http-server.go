package qbin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

func escapeHTML(content string) string {
	return strings.Replace(strings.Replace(strings.Replace(content, "&", "&amp;", -1), "<", "&lt;", -1), ">", "&gt;", -1)
}

func rootReplace(content string) string {
	return strings.Replace(strings.Replace(content, "$$path$$", Config["path"], -1), "$$root$$", Config["root"], -1)
}

////////////
// Routes //
////////////

func staticRoute(body string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		// Get static file
		res.Header().Add("Content-Type", mime.TypeByExtension(filepath.Ext(req.URL.Path))+"; charset=utf-8")
		fmt.Fprintf(res, "%s", body)
	}
}

func staticAliasRoute(path string, manipulateSource func(string) string, manipulateResult func(*http.Request, string) string) func(http.ResponseWriter, *http.Request) {
	body, err := loadStaticFile(Config["frontend-path"]+path, path, false)
	if manipulateSource != nil {
		body = manipulateSource(body)
	}
	if err == nil {
		return func(res http.ResponseWriter, req *http.Request) {
			// Get index page, replace content with forked content eventually
			if manipulateResult != nil {
				body = manipulateResult(req, body)
			}
			res.Header().Add("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(res, "%s", body)
		}
	}

	return internalErrorRoute
}

func uploadRoute(res http.ResponseWriter, req *http.Request) {
	// TODO: Upload document
	fmt.Fprintf(res, "Welcome to the upload page!")
}

func rawDocumentRoute(res http.ResponseWriter, req *http.Request) {
	// TODO: Get raw content
	fmt.Fprintf(res, `console.log("Hello <World>");`+"\n")
}

func documentRoute() func(http.ResponseWriter, *http.Request) {
	body, err := loadStaticFile(Config["frontend-path"]+"/output.html", "/output.html", false)
	if err == nil {
		return func(res http.ResponseWriter, req *http.Request) {
			// TODO: Check for curl/wget requests and return raw document

			content := `console.log("Hello <World>");` + "\n"
			content = `<pre class="no-linenumber-padding"><code class="language-javascript line-numbers">` + escapeHTML(content) + `</code></pre>`
			returnBody := strings.Replace(body, "$$id$$", strings.TrimPrefix(req.URL.Path, "/"), -1)
			returnBody = strings.Replace(returnBody, "$$syntax$$", "javascript", -1)
			returnBody = strings.Replace(returnBody, "$$creation$$", "2017-01-01 12:34", -1)
			returnBody = strings.Replace(returnBody, "$$expiration$$", "2017-05-01 12:34", -1)
			returnBody = strings.Replace(returnBody, "$$expiration-remaining$$", "2 days left", -1)
			returnBody = strings.Replace(returnBody, "$$content$$", content, -1)

			res.Header().Add("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(res, "%s", returnBody)
		}
	}

	return internalErrorRoute
}

func forkDocumentRoute() func(http.ResponseWriter, *http.Request) {
	return staticAliasRoute("/index.html", func(body string) string {
		return regexp.MustCompile("\\$\\$/?if_fork\\$\\$").ReplaceAllString(
			regexp.MustCompile("\\$\\$!if_fork\\$\\$(?U:.*\\$\\$/if_fork\\$\\$)").ReplaceAllString(body, ""), "")
	}, func(req *http.Request, body string) string {
		content := `console.log("Hello <World>");` + "\n"
		return strings.Replace(body, "></textarea>", ">"+escapeHTML(content)+"</textarea>", 1)
	})
}

func internalErrorRoute(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(500)
	fmt.Fprint(res, "Oh no, the server is broken! ಠ_ಠ\nYou should try again in a few minutes, there's probably a desperate admin running around somewhere already trying to fix it.")
}

func notFoundRoute(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(404)
	fmt.Fprint(res, "Oops, seems like there's nothing here! ¯\\_(ツ)_/¯\nMaybe the document is expired or has been removed.")
}

//////////////////
// Static Files //
//////////////////

var staticFileExceptions = []string{
	"index.html",
	"output.html",
	"report.html",
	"api.html",
	"guidelines.html",
	"Makefile",
	"README.md",
}

func isStaticFileExempt(filename string) bool {
	if strings.HasSuffix(filename, ".styl") {
		return true
	}
	for _, exemptFilename := range staticFileExceptions {
		if filename == exemptFilename {
			return true
		}
	}
	return false
}

func loadStaticFile(path string, webPath string, checkExceptions bool) (string, error) {
	if checkExceptions && isStaticFileExempt(filepath.Base(path)) {
		log.Debugf("Exempted static file '%s'.", webPath)
		return "", errors.New("file is exempt")
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("Couldn't read static file '%s' for the following reason: %s", webPath, err)
		return "", err
	}

	log.Debugf("Added static file '%s'.", webPath)
	return rootReplace(string(content)), nil
}

func addStaticRoute(r *mux.Router) func(string, os.FileInfo, error) error {
	return func(path string, f os.FileInfo, err error) error {
		if err == nil {
			webPath := strings.TrimPrefix(path, Config["frontend-path"])
			if !strings.HasPrefix(webPath, "/") {
				webPath = "/" + webPath
			}

			if !f.IsDir() && !strings.Contains(webPath, "/.") {

				// Read & serve file
				content, err := loadStaticFile(path, webPath, true)
				if err == nil {
					r.HandleFunc(webPath, staticRoute(content)).Methods("GET")
				}

			} else if isStaticFileExempt(f.Name()) {
				log.Debugf("Exempted static file '%s'.", webPath)
			}
		}
		return nil
	}
}

///////////////////////////
// Server Initialization //
///////////////////////////

// StartHTTPServer starts the HTTP server which is responsible for the frontend and the HTTP API.
func StartHTTPServer() {
	r := mux.NewRouter()

	// Upload function
	r.HandleFunc("/", uploadRoute).Methods("POST")

	// Static aliased HTML files
	r.HandleFunc("/", staticAliasRoute("/index.html", func(body string) string {
		return regexp.MustCompile("\\$\\$[/!]if_fork\\$\\$").ReplaceAllString(
			regexp.MustCompile("\\$\\$if_fork\\$\\$(?U:.*\\$\\$/if_fork\\$\\$)").ReplaceAllString(body, ""), "")
	}, nil)).Methods("GET")
	r.HandleFunc("/api", staticAliasRoute("/api.html", nil, nil)).Methods("GET")
	r.HandleFunc("/guidelines", staticAliasRoute("/guidelines.html", nil, nil)).Methods("GET")

	// Static files
	filepath.Walk(Config["frontend-path"], addStaticRoute(r))

	// Documents
	r.HandleFunc("/{document}", documentRoute()).Methods("GET")
	r.HandleFunc("/{document}/raw", rawDocumentRoute).Methods("GET")
	r.HandleFunc("/{document}/fork", forkDocumentRoute()).Methods("GET")
	r.HandleFunc("/{document}/report", staticAliasRoute("/report.html", nil, func(req *http.Request, body string) string {
		return strings.Replace(body, "$$id$$", strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/report"), -1)
	})).Methods("GET")

	// 404 error page
	r.PathPrefix("/").HandlerFunc(notFoundRoute)

	log.Noticef("HTTP server now listening on %s, you should be able to reach it at %s", Config["http"], Config["root"]+Config["path"])
	go http.ListenAndServe(Config["http"], r)
}
