package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//TODO Use logrus for logstash
var logjson = logrus.New()

type handlerInConn struct {
	schemeIn  string
	hostIn    string
	portIn    int
	schemeOut string
	hostOut   string
	portOut   int
}

var myHandler handlerInConn

func init() {
	log.Println("Init...")
	// v:=viper.New()
	viper.SetDefault("listen.scheme", "http")
	viper.SetDefault("listen.interface", "")
	viper.SetDefault("listen.port", 8089)
	viper.SetDefault("forward.scheme", "http")
	viper.SetDefault("forward.interface", "127.0.0.1")
	viper.SetDefault("forward.port", 1323)
	viper.SetConfigFile("originproxy.yml")
	// viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/originproxy/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
	}
	myHandler.schemeIn = viper.GetString("listen.scheme")
	myHandler.hostIn = viper.GetString("listen.interface")
	myHandler.portIn = viper.GetInt("listen.port")
	myHandler.schemeOut = viper.GetString("forward.scheme")
	myHandler.hostOut = viper.GetString("forward.interface")
	myHandler.portOut = viper.GetInt("forward.port")
}

func main() {
	logjson.SetFormatter(&logrus.JSONFormatter{})
	logjson.Println("Start...")
	inConnAddress := myHandler.hostIn + ":" + strconv.Itoa(myHandler.portIn)
	outConnAddress := myHandler.schemeOut + "://" + myHandler.hostOut + ":" + strconv.Itoa(myHandler.portOut)
	log.Println(myHandler.schemeIn + "://" + inConnAddress + " -> " + outConnAddress)
	serverIn := &http.Server{
		Addr:    inConnAddress,
		Handler: &myHandler,
	}
	logjson.Fatal(serverIn.ListenAndServe())
}

func searchOrigin(s string) bool {
	origins := viper.GetStringSlice("origin")
	sort.Strings(origins)
	index := sort.SearchStrings(origins, s)
	if sort.SearchStrings(origins, s) == len(origins) || origins[index] != s {
		return false
	}
	return true
}

func (hin *handlerInConn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	log.Println(origin)
	log.Println(r.RemoteAddr, r.Host, r.Method, r.RequestURI)
	if len(origin) == 0 || !searchOrigin(origin) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}
	outAddress := hin.schemeOut + "://" + hin.hostOut + ":" + strconv.Itoa(hin.portOut)
	outURL, err := url.Parse(outAddress)
	if err != nil {
		log.Println(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(outURL)
	r.URL.Host = outURL.Host
	r.URL.Scheme = outURL.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = outURL.Host
	proxy.ServeHTTP(w, r)
}

