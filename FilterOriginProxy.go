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
	portIn    int
	hostIn    string
	schemeIn  string
	portOut   int
	hostOut   string
	schemeOut string
}

var myHandler handlerInConn

func init() {
	log.Println("Init...")
	// v:=viper.New()
	viper.SetDefault("scheme_listen", "http")
	viper.SetDefault("scheme_to", "http")
	viper.SetDefault("interface_listen", "0.0.0.0")
	viper.SetDefault("interface_to", "127.0.0.1")
	viper.SetDefault("port_listen", 8089)
	viper.SetDefault("port_to", 1323)
	viper.SetConfigFile("originproxy.yml")
	// viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/originproxy/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
	}
	myHandler.schemeIn = viper.GetString("scheme_listen")
	myHandler.hostIn = viper.GetString("interface_listen")
	myHandler.portIn = viper.GetInt("port_listen")
	myHandler.schemeOut = viper.GetString("scheme_to")
	myHandler.hostOut = viper.GetString("interface_to")
	myHandler.portOut = viper.GetInt("port_to")
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
