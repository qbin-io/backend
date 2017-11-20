package qbinHTTP

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/qbin-io/backend"
)

// setupRoutes will set up a mux Router to provide the routes used by the qbin frontend and API.
func setupRoutes(r *mux.Router) {
	// Upload function
	r.HandleFunc("/", uploadRoute).Methods("POST")

	// Static aliased HTML files
	r.HandleFunc("/", advancedStaticRoute(config.FrontendPath, "/index.html", routeOptions{
		ignoreExceptions: true,
		modifySource: func(body *string) {
			replaceBlockVariable(body, "if_fork", false)
		},
	})).Methods("GET")
	r.HandleFunc("/guidelines", staticRoute(config.FrontendPath, "/guidelines.html", true)).Methods("GET")

	// Static files
	qbin.Log.Debugf("Including static files from: %s", config.FrontendPath)
	addStaticDirectory(config.FrontendPath, "/", r)

	// Documents
	r.HandleFunc("/{document}", documentRoute()).Methods("GET")
	r.HandleFunc("/{document}/raw", rawDocumentRoute).Methods("GET")
	r.HandleFunc("/{document}/fork", forkDocumentRoute()).Methods("GET")
	r.HandleFunc("/{document}/report", advancedStaticRoute(config.FrontendPath, "/report.html", routeOptions{
		ignoreExceptions: true,
		modifyResult: func(res http.ResponseWriter, req *http.Request, body *string) error {
			replaceVariable(body, "id", strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/report"))
			return nil
		},
	})).Methods("GET")

	// 404 error page
	r.PathPrefix("/").HandlerFunc(notFoundRoute)
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
	return advancedStaticRoute(config.FrontendPath, "/output.html", routeOptions{
		ignoreExceptions: true,
		modifyResult: func(res http.ResponseWriter, req *http.Request, body *string) error {
			// TODO: Check for curl/wget requests and return raw document
			if false {
				// Does this work?!
				rawDocumentRoute(res, req)
				return errors.New("Serving for curl")
			}

			content := `console.log("Hello <World>");` + "\n"
			content = `<pre class="no-linenumber-padding"><code class="language-javascript line-numbers">` + escapeHTML(content) + `</code></pre>`
			replaceVariable(body, "id", strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/fork"))
			replaceVariable(body, "syntax", "javascript")
			replaceVariable(body, "creation", "2017-01-01 12:34 (UTC)")
			replaceVariable(body, "expiration", "2017-05-01 12:34 (UTC)")
			replaceVariable(body, "expiration-remaining", "2 days left")
			replaceVariable(body, "content", content)

			return nil
		},
	})
}

func forkDocumentRoute() func(http.ResponseWriter, *http.Request) {
	return advancedStaticRoute(config.FrontendPath, "/index.html", routeOptions{
		ignoreExceptions: true,
		modifySource: func(body *string) {
			replaceBlockVariable(body, "if_fork", true)
		},
		modifyResult: func(res http.ResponseWriter, req *http.Request, body *string) error {
			content := `console.log("Hello <World>");` + "\n"
			replaceVariable(body, "id", strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/fork"))
			replaceVariable(body, "syntax", "javascript")
			replaceVariable(body, "creation", "2017-01-01 12:34 (UTC)")
			replaceVariable(body, "expiration", "2017-05-01 12:34 (UTC)")
			replaceVariable(body, "expiration-remaining", "2 days left")
			replaceVariable(body, "content", escapeHTML(content))

			return nil
		},
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
