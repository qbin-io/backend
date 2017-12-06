package qbinTCP

import (
	"net"
	"time"

	"github.com/qbin-io/backend"
)

var froot string

// StartTCP launches the TCP server which is responsible for the TCP API.
//listen <- given tcp port
func StartTCP(listen string, root string) {
	froot = root

	//reading server TCP-Address
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listen)
	checkErr(err)

	//starting listener on TCP-Address
	listenerTCP, err := net.ListenTCP("tcp", tcpAddr) //not sure why to use tcpAddr instead of listen
	checkErr(err)

	qbin.Log.Noticef("TCP server starting on %s, you should be able to reach it at %s", listen, froot)

	//waiting for incoming connections
	for {
		connTCP, err := listenerTCP.AcceptTCP()
		if err != nil {
			qbin.Log.Error("TCP Server Error:", err.Error())
			continue
		}

		go handleClient(connTCP)
	}
}

//log Error if one occures
func checkErr(err error) {
	if err != nil {
		qbin.Log.Errorf("TCP Server Error: %s", err.Error())
	}
}

//handle incomming TCP Connection
func handleClient(connTCP net.Conn) {
	//set timeout to 2 minutes, limit requestlength to 1MB, close connection on programm close
	connTCP.SetReadDeadline(time.Now().Add(2 * time.Minute))
	request := make([]byte, 1024)
	defer connTCP.Close()

	for {
		msgLength, err := connTCP.Read(request)
		if err != nil {
			qbin.Log.Errorf("TCP Read Error: %s", err.Error())
			break
		}

		msg := string(request[:msgLength])
		handleMsgProcessing(connTCP, msg)

		connTCP.Close()
		break
	}
}

func handleMsgProcessing(connTCP net.Conn, msg string) {
	//length = 0 -> user already closed connection
	if len(msg) == 0 {
		return
	}
	//there is a msg, so we process it
	//sendIP := connTCP.RemoteAddr().String
	//qbin.Log.Notice("You received something via TCP")
	//qbin.Log.Notice(msg)

	exp, err := qbin.ParseExpiration("14d")
	checkErr(err)

	doc := qbin.Document{
		Content:    msg,
		Syntax:     "",
		Expiration: exp,
		Address:    "::ffff:127.0.0.1",
	}

	err = qbin.Store(&doc)
	checkErr(err) //Todo: User err msg

	reply := froot + "/" + doc.ID + "\n"

	connTCP.Write([]byte(reply))
}
