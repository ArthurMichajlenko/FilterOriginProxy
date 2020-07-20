package main

import (
	// "io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// type handlerInConn struct{}
type handlerInConn struct {
	port int
	host string
}

func (hin *handlerInConn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	log.Println(err)
	// }
	log.Println(origin)
	http.Redirect(w, r, "http://localhost:1323"+r.RequestURI, http.StatusFound)
}

func main() {
	//TODO Use logrus for logstash
	logjson := logrus.New()
	logjson.SetFormatter(&logrus.JSONFormatter{})
	logjson.Println("Start...")
	// Create inbound connection
	var myHandler handlerInConn
	myHandler.port = 8089
	inConnAddress := myHandler.host + ":" + strconv.Itoa(myHandler.port)
	log.Println(inConnAddress)
	serverIn := &http.Server{
		Addr:    inConnAddress,
		Handler: &myHandler,
	}
	logjson.Fatal(serverIn.ListenAndServe())
}
