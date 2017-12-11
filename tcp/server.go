package qbinTCP

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/qbin-io/backend"
)

// StartTCP launches the TCP server which is responsible for the TCP API.
//listen <- given tcp port
func StartTCP(listen string, root string) {
	root = strings.TrimSuffix(root, "/")

	//reading server TCP-Address
	qbin.Log.Debug("Resolving TCP...")
	tcpAddr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		qbin.Log.Errorf("There was an Error resolving the TCPAddress \n%s\n The Service will be stoped now!", err)
		return
	}

	//starting listener on TCP-Address
	qbin.Log.Debug("Starting TCP-listener")
	listenerTCP, err := net.ListenTCP("tcp", tcpAddr) //not sure why to use tcpAddr instead of listen
	if err != nil {
		qbin.Log.Errorf("There was an Error starting the TCP-listener \n%s\n The Service will be stoped now!", err)
		return
	}

	qbin.Log.Noticef("TCP server starting on %s, you should be able to reach it at %s%s", listen, getTCProot(root), listen)

	//waiting for incoming connections
	for {
		connTCP, err := listenerTCP.AcceptTCP()
		if err != nil {
			qbin.Log.Errorf("TCP Server Error: %s", err.Error())
			continue
		}

		go handleClient(connTCP, root)
	}
}

func getTCProot(root string) string {
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
func handleClient(connTCP net.Conn, root string) {
	qbin.Log.Debug("Handeling incomming connection...")
	//set timeout to 1.5 seconds, limit requestlength to 1MB, close connection on programm close
	defer connTCP.Close()
	msg := ""

	for {
		var b = make([]byte, 8)
		connTCP.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
		i, err := connTCP.Read(b)
		if err != nil && err != io.EOF && !(strings.Contains(err.Error(), "i/o timeout")) {
			qbin.Log.Errorf("TCP-read error: %s", err)
			connTCP.Write([]byte("Ups, something went wrong. \n"))
			return
		}
		msg += string(b[:i])

		if len(msg) > qbin.MaxFilesize {
			qbin.Log.Error("The recived file is to big! refusing connection")
			connTCP.Write([]byte("Your filesize is over 1M. \n"))
			return
		}

		if err != nil && strings.HasSuffix(err.Error(), "i/o timeout") {
			qbin.Log.Errorf("The connection timed out while reading: %s", err)
			handleMsgProcessing(connTCP, msg, root)
			return
		}

		if err != nil && err == io.EOF {
			handleMsgProcessing(connTCP, msg, root)
			return
		}

	}

}

func handleMsgProcessing(connTCP net.Conn, msg string, root string) {
	//length = 0 -> user already closed connection
	if len(msg) == 0 {
		qbin.Log.Debug("The TCP message was empty")
		connTCP.Write([]byte("We received an empty message. \n"))
		return
	}

	//there is a msg, so we process it
	exp, err := qbin.ParseExpiration("14d")
	if err != nil {
		qbin.Log.Errorf("There was an expiration parsing error: %s", err)
		connTCP.Write([]byte("Ups, something went wrong.\n"))
		return
	}

	doc := qbin.Document{
		Content:    msg,
		Syntax:     "",
		Expiration: exp,
		Address:    connTCP.RemoteAddr().String(),
	}

	err = qbin.Store(&doc)
	if err != nil {
		if err.Error() == "file contains 0x00 bytes" {
			qbin.Log.Errorf("There was an error storing the TCP input: %s \nThe uploaded binary file is not supported.\n", err)
			connTCP.Write([]byte("You are trying to upload a binary file, which is not supported.\n"))
		} else {
			qbin.Log.Errorf("There was an error storing the TCP input: %s", err)
			connTCP.Write([]byte("Ups, something went wrong. \n"))
		}
		return
	}

	reply := root + "/" + doc.ID + "\n"

	connTCP.Write([]byte(reply))
	qbin.Log.Debug("TCP document was saved")
}
