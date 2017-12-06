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
	tcpAddr, err := net.ResolveTCPAddr("tcp", listen)
	checkErr(err)

	//starting listener on TCP-Address
	listenerTCP, err := net.ListenTCP("tcp", tcpAddr) //not sure why to use tcpAddr instead of listen
	checkErr(err)

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

//log Error if one occures
func checkErr(err error) {
	if err != nil {
		qbin.Log.Errorf("TCP Server Error: %s", err.Error())
	}
}

//handle incomming TCP Connection
func handleClient(connTCP net.Conn, root string) {
	//set timeout to 5 seconds, limit requestlength to 1MB, close connection on programm close
	connTCP.SetReadDeadline(time.Now().Add(5 * time.Second))
	result := make([]byte, 1024)
	defer connTCP.Close()

	result, err := ioutil.ReadAll(connTCP)
	checkErr(err)
	handleMsgProcessing(connTCP, string(result), root)

}

func handleMsgProcessing(connTCP net.Conn, msg string, root string) {
	//length = 0 -> user already closed connection
	if len(msg) == 0 {
		return
	}

	//there is a msg, so we process it
	exp, err := qbin.ParseExpiration("14d")
	checkErr(err)

	doc := qbin.Document{
		Content:    msg,
		Syntax:     "",
		Expiration: exp,
		Address:    connTCP.RemoteAddr().String(),
	}

	err = qbin.Store(&doc)
	if err != nil {
		checkErr(err)
		connTCP.Write([]byte("Ups, something went wrong."))
		return
	}

	reply := root + "/" + doc.ID + "\n"

	connTCP.Write([]byte(reply))
}
