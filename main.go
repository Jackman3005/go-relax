package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

var app CloudFoundryApp

func main() {
	config := loadConfiguration()
	app = CloudFoundryApp{
		CFAppGUID: config.CFAppGUID,
	}
	app.printSummary()

	go func() {
		for true {
			time.Sleep(15 * time.Second)
			app.updateState()
		}
	}()

	// start server
	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	if app.CurrentState != APP_RUNNING {
		fmt.Println("App is currently unavailable")
		app.start()
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
