package qbinTCP

import (
	"io/ioutil"
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
	//set timeout to 5 seconds, limit requestlength to 1MB, close connection on programm close
	connTCP.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
	result := make([]byte, 1024)
	defer connTCP.Close()

	result, err := ioutil.ReadAll(connTCP)
	if err != nil {
		qbin.Log.Errorf("TCP-read error: %s", err)
		connTCP.Write([]byte("Ups, something went wrong."))
		return
	}
	handleMsgProcessing(connTCP, string(result), root)

}

func handleMsgProcessing(connTCP net.Conn, msg string, root string) {
	//length = 0 -> user already closed connection
	if len(msg) == 0 {
		qbin.Log.Debug("The TCP message was empty")
		return
	}

	//there is a msg, so we process it
	exp, err := qbin.ParseExpiration("14d")
	if err != nil {
		qbin.Log.Errorf("There was an expiration parsing error: %s", err)
	}

	doc := qbin.Document{
		Content:    msg,
		Syntax:     "",
		Expiration: exp,
		Address:    connTCP.RemoteAddr().String(),
	}

	err = qbin.Store(&doc)
	if err != nil {
		qbin.Log.Errorf("There was an error storing the TCP input: %s", err)
		connTCP.Write([]byte("Ups, something went wrong."))
		return
	}

	reply := root + "/" + doc.ID + "\n"

	connTCP.Write([]byte(reply))
	qbin.Log.Debug("TCP document was saved")
}
