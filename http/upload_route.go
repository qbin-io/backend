package qbinHTTP

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/qbin-io/backend"
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
	sizeExceeded := false

	// Parse form and get content
	req.Body = http.MaxBytesReader(res, req.Body, qbin.MaxFilesize+1024) // MaxFilesize + 1KB metadata
	contentType := strings.Split(strings.Replace(strings.ToLower(req.Header.Get("Content-Type")), " ", "", -1), ";")[0]

	// Get the document, however the request is formatted
	if req.Method == "POST" && contentType == "application/x-www-form-urlencoded" {
		// Parse form
		err = req.ParseForm()
		if err != nil && err.Error() == "http: request body too large" {
			sizeExceeded = true
		} else if uploadError("req.ParseForm()", err, res, req) {
			return
		} else {

			// Get document
			doc.Content = req.PostFormValue("Q")

		}
	} else if req.Method == "POST" && contentType == "multipart/form-data" {
		// Parse form
		err = req.ParseMultipartForm(qbin.MaxFilesize + 1024)
		if err != nil && err.Error() == "http: request body too large" {
			sizeExceeded = true
		} else if uploadError("req.ParseMultipartForm()", err, res, req) {
			return
		} else {

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

		}
	} else { // PUT or POST with non-form
		// Read document
		content, err := ioutil.ReadAll(req.Body)
		if err != nil && err.Error() == "http: request body too large" {
			sizeExceeded = true
		} else if uploadError("ioutil.ReadAll()", err, res, req) {
			return
		} else {

			doc.Content = string(content)

		}
	}

	// Check exact filesize
	if sizeExceeded || len(doc.Content) > qbin.MaxFilesize {
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
		return
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

	// Volatile documents grant 1 view to the uploader, but the uploaded won't view a document when not redirected
	if !redirect {
		doc.Views = 1
	}

	err = qbin.Store(&doc)
	if err != nil && err.Error() == "file contains 0x00 bytes" {
		res.WriteHeader(400)
		fmt.Fprintf(res, "You are trying to upload a binary file, which is not supported.\n")
		return
	} else if err != nil && strings.HasPrefix(err.Error(), "spam: ") {
		res.WriteHeader(400)
		fmt.Fprintf(res, "Your file got caught in the spam filter.\nReason: "+strings.TrimPrefix(err.Error(), "spam: ")+"\n")
		return
	} else if uploadError("qbin.Store()", err, res, req) {
		return
	}

	// Redirect or return URL
	if redirect {
		res.Header().Set("Location", config.Root+"/"+doc.ID)
		res.WriteHeader(302)
	}
	fmt.Fprintf(res, config.Root+"/"+doc.ID+"\n")
}
