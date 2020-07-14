package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	//TODO Use logrus for logstash
	logjson := logrus.New()
	logjson.SetFormatter(&logrus.JSONFormatter{})
	// Create inbound connection
	connIn, err := net.Listen("tcp", ":8089")
	if err != nil {
		logjson.Logln(logrus.ErrorLevel, err)
	}
	log.Println("Listener started")
	defer func() {
		connIn.Close()
		logjson.Logln(logrus.DebugLevel, "Listener closed")
	}()
	for {
		conn, err := connIn.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleConnectionIn(conn)
	}
}

func handleConnectionIn(conn net.Conn) {
	log.Println("Handling new connection...")
	reqHeader := ""
	origin := ""
	defer func() {
		log.Println("Closing connection...")
		conn.Close()
	}()
	timeoutDuration := 5 * time.Second
	bufReader := bufio.NewReader(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		headItem, err := bufReader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		reqHeader = reqHeader + headItem
		if strings.HasPrefix(headItem, "Origin:") {
			origin = strings.SplitAfter(headItem, ": ")[1]
		}
		if headItem == "\r\n" {
			break
		}
	}
	// Check if Headers contains "Origin"
	if len(origin) == 0 {
		log.Println(`No "Origin" in header`)
		return
	}
	log.Println("End of header")
	fmt.Printf("%s", origin)
	fmt.Printf("%v", reqHeader)
	connOut, err := net.Dial("tcp", "localhost:1323")
	if err != nil {
		log.Println(err)
		return
	}
	// conn, connOut = net.Pipe()
	fmt.Fprintf(connOut, "%v", reqHeader)
	var b []byte
	count, err := connOut.Read(b)
	if err != nil {
		log.Println(err)
	}
	log.Println(count)
	log.Println(b)
	// fmt.Fprintf(conn, "%v", reqHeader)
}
