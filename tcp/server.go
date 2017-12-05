package qbinTCP

import (
	"net"
	"time"

	"github.com/qbin-io/backend"
)

// StartTCP launches the TCP server which is responsible for the TCP API.
//listen <- given tcp port
func StartTCP(listen string, root string) {
	qbin.Log.Warning("TCP server is work in Progess")

	//reading server TCP-Adress and printing to log
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listen)
	checkErr(err)
	qbin.Log.Noticef("TCP listener on Port %s.", tcpAddr)

	//starting listener on previously mentiond port
	listenerTCP, err := net.ListenTCP("tcp", tcpAddr) //not sure why to use tcpAddr instead of listen
	checkErr(err)

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
		qbin.Log.Error("TCP Server Error: ", err.Error())
	}
}

//handle incomming TCP Connection
func handleClient(connTCP net.Conn) {
	//set timeout, limit requestlength to 128B, close connection on close
	connTCP.SetReadDeadline(time.Now().Add(2 * time.Minute))
	request := make([]byte, 128)
	defer connTCP.Close()

	for {
		msgLength, err := connTCP.Read(request)
		if err != nil {
			qbin.Log.Error("TCP Read Error:", err.Error())
			break
		}

		msg := string(request[:msgLength])
		qbin.Log.Notice(msgLength)
		qbin.Log.Notice(len(msg))
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
	qbin.Log.Notice("You recived something via TCP")
	qbin.Log.Notice(msg)
	connTCP.Write([]byte("We got your message \n"))

}
