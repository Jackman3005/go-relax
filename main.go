package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	printAppSummary()
	//turnAppOff()

	// start server
	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
	//var j int
	//_ = j
	//
	//for i:=0; i < 10000; i++ {
	//	prime:=true
	//	for k:=i-1; k > 1; k-- {
	//		if i%k == 0 {
	//			prime = false
	//			break
	//		}
	//	}
	//	if prime {
	//		fmt.Println(i)
	//	}
	//}
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	if !canAppReceiveRequests() {
		fmt.Println("App is currently unavailable")
		turnAppOn()
	}

	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url := getEnvOrDefault("TARGET_API_URI", "http://localhost:8080")
	serveReverseProxy(url, res, req)
}

func getListenAddress() string {
	port := getEnvOrDefault("PORT", "1337")
	return ":" + port
}

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
