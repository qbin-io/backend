package qbinHTTP

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	qbin "github.com/qbin-io/backend"
)

func uploadError(during string, err error, res http.ResponseWriter, req *http.Request) bool {
	if err == nil {
		return false
	}
	qbin.Log.Errorf("Upload error during %s: %s", during, err)
	internalErrorRoute(res, req)
	return true
}

func uploadRoute(res http.ResponseWriter, req *http.Request) {
	var err error

	doc := qbin.Document{Address: req.RemoteAddr}
	exp := "14d"
	redirect := false

	// Parse form and get content
	// TODO: custom error message for huge bodies
	req.Body = http.MaxBytesReader(res, req.Body, qbin.MaxFilesize)
	contentType := strings.Split(strings.Replace(strings.ToLower(req.Header.Get("Content-Type")), " ", "", -1), ";")[0]

	// Get the document, however the request is formatted
	if req.Method == "POST" && contentType == "application/x-www-form-urlencoded" {
		// Parse form
		err = req.ParseForm()
		if uploadError("req.ParseForm()", err, res, req) {
			return
		}

		// Get document
		doc.Content = req.PostFormValue("Q")
	} else if req.Method == "POST" && contentType == "multipart/form-data" {
		// Parse form
		err = req.ParseMultipartForm(qbin.MaxFilesize)
		if uploadError("req.ParseMultipartForm()", err, res, req) {
			return
		}

		// Get document
		doc.Content = req.PostFormValue("Q")
		if doc.Content == "" { // Oh no, it's a file!
			// Get file
			file, _, err := req.FormFile("Q")
			if err != nil && err.Error() == "http: no such file" {
				res.WriteHeader(400)
				fmt.Fprintf(res, "The document can't be empty.\n")
				return
			} else if uploadError("req.FormFile()", err, res, req) {
				return
			}

			// Read document
			content, err := ioutil.ReadAll(file)
			if uploadError("ioutil.ReadAll()", err, res, req) {
				return
			}
			doc.Content = string(content)
		}
	} else { // PUT or POST with non-form
		// TODO: test for huge bodies!
		// Read document
		content, err := ioutil.ReadAll(req.Body)
		if uploadError("ioutil.ReadAll()", err, res, req) {
			return
		}
		doc.Content = string(content)
	}

	// Check exact filesize
	if len(doc.Content) > qbin.MaxFilesize {
		res.WriteHeader(413)
		fmt.Fprintf(res, "Maximum document size exceeded.\n")
		return
	}
	if len(strings.TrimSpace(doc.Content)) < 1 {
		res.WriteHeader(400)
		fmt.Fprintf(res, "The document can't be empty.\n")
		return
	}

	// Read metadata
	if req.Header.Get("S") != "" {
		doc.Syntax = req.Header.Get("S")
	} else if req.FormValue("S") != "" {
		doc.Syntax = req.FormValue("S")
	}
	doc.Syntax = qbin.ParseSyntax(doc.Syntax)
	if !qbin.SyntaxExists(doc.Syntax) {
		res.WriteHeader(400)
		fmt.Fprintf(res, "Invalid syntax name.\n")
	}

	if req.Header.Get("R") != "" || req.FormValue("R") != "" {
		redirect = true
	}

	if req.Header.Get("E") != "" {
		exp = req.Header.Get("E")
	} else if req.FormValue("E") != "" {
		exp = req.FormValue("E")
	}

	doc.Expiration, err = qbin.ParseExpiration(exp)
	if err != nil {
		res.WriteHeader(400)
		fmt.Fprintf(res, "Invalid expiration.\n")
		return
	}

	// TODO: Store
	err = qbin.Store(&doc)
	if uploadError("qbin.Store()", err, res, req) {
		return
	}

	// Redirect or return URL
	if redirect {
		res.Header().Set("Location", config.Root+"/"+doc.ID)
		res.WriteHeader(302)
	}
	fmt.Fprintf(res, config.Root+"/"+doc.ID+"\n")
}
