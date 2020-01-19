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
	app = NewCloudFoundryApp(config.CFAppGUID)
	app.printSummary()

	go func() {
		for true {
			time.Sleep(5 * time.Second)
			var previousState = app.CurrentState
			app.updateState()
			if previousState != app.CurrentState {
				fmt.Println("App state changed from " + previousState + " to " + app.CurrentState)
			}
		}
	}()
	go func() {
		for true {
			time.Sleep(15 * time.Second)
			timeSinceLastRequest := time.Now().Sub(app.LastRequestTimestamp)
			inactivityThreshold, _ := time.ParseDuration(config.InactivityThreshold)
			if app.CurrentState == APP_RUNNING && timeSinceLastRequest > inactivityThreshold {
				fmt.Println("The app has not received any requests for at least ", inactivityThreshold, ". \nShutting down app...")
				fmt.Println("Last request received at: ", app.LastRequestTimestamp)
				app.stop()
			}
		}
	}()

	// start server
	http.HandleFunc("/", handleRequestAndRedirect(config.TargetURL))
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	app.LastRequestTimestamp = time.Now()
	if app.CurrentState != APP_RUNNING {
		app.start()
	}

	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	app.LastRequestTimestamp = time.Now()
	proxy.ServeHTTP(res, req)
}

func handleRequestAndRedirect(url string) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		serveReverseProxy(url, res, req)
	}
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
