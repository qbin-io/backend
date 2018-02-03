package qbinTCP

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/qbin-io/backend"
)

// StartTCP launches the TCP server which is responsible for the TCP API.
func StartTCP(listen string, root string) {
	qbin.Log.Debug("Initializing TCP server...")
	root = strings.TrimSuffix(root, "/")

	// Resolve TCP listen address
	tcpAddr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		qbin.Log.Criticalf("TCP resolve error: %s", err)
		panic(err)
	}

	// Listen
	qbin.Log.Noticef("TCP server starting on %s, you should be able to reach it at %s%s", listen, getRoot(root), listen)
	listenerTCP, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		qbin.Log.Criticalf("TCP server error: %s", err)
		panic(err)
	}

	// Wait for incoming connections
	for {
		conn, err := listenerTCP.AcceptTCP()
		if err != nil {
			qbin.Log.Errorf("TCP server error: %s", err)
			continue
		}

		go handleClient(conn, root)
	}
}

func getRoot(root string) string {
	if strings.Contains(root, "://") {
		root = root[strings.Index(root, "://")+3:]
	}
	for strings.Contains(root, "/") {
		root = root[0:strings.Index(root, "/")]
	}
	if strings.Contains(root, ":") {
		root = root[0:strings.Index(root, ":")]
	}
	return root
}

//handle incomming TCP Connection
func handleClient(conn net.Conn, root string) {
	// Set timeout to 1.5 seconds, limit requestlength to 1MB, close connection on programm close
	defer conn.Close()
	msg := ""

	for {
		// Set timeout to 1.5 seconds on every read
		conn.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))

		var b = make([]byte, 8)
		i, err := conn.Read(b)
		if err != nil && err != io.EOF && !strings.Contains(err.Error(), "i/o timeout") {
			qbin.Log.Errorf("TCP read error: %s", err)
			conn.Write([]byte("An error occured, please try again.\n"))
			return
		}
		msg += string(b[:i])

		if len(msg) > qbin.MaxFilesize {
			conn.Write([]byte("Maximum document size exceeded.\n"))
			return
		}

		if err != nil {
			handleMsgProcessing(conn, msg, root)
			return
		}
	}
}

var defaultExpiration, _ = qbin.ParseExpiration("14d")

func handleMsgProcessing(conn net.Conn, msg string, root string) {
	//length = 0 -> user already closed connection
	if len(msg) == 0 {
		conn.Write([]byte("The document can't be empty.\n"))
		return
	}

	// Create and store document
	doc := qbin.Document{
		Content:    msg,
		Syntax:     "",
		Expiration: defaultExpiration,
		Address:    conn.RemoteAddr().String(),
		Views:      1,
	}

	err := qbin.Store(&doc)
	if err != nil {
		if err.Error() == "file contains 0x00 bytes" {
			conn.Write([]byte("You are trying to upload a binary file, which is not supported.\n"))
		} else if strings.HasPrefix(err.Error(), "spam: ") {
			conn.Write([]byte("Your file got caught in the spam filter.\nReason: " + strings.TrimPrefix(err.Error(), "spam: ") + "\n"))
		} else {
			qbin.Log.Errorf("TCP API error: %s", err)
			conn.Write([]byte("An error occured, please try again.\n"))
		}
		return
	}

	// Send URL back
	reply := root + "/" + doc.ID + "\n"
	conn.Write([]byte(reply))
}
