package main

import (
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient"
	"strconv"
	"time"
)

var cfClient, _ = cfclient.NewClient(&cfclient.Config{
	ApiAddress: "https://api.run.pivotal.io",
	Username:   "jackman3000@gmail.com",
	Password:   "1hGGrmZ3bl%uEvgb*2@bNy&92jO6BH",
})
var appGUID = "a3d07cd1-5d0c-4cfa-8faa-3ad7214be7a1" // cf app api --guid

var appCanReceiveRequests bool = false

func printAppSummary() {
	app, _ := cfClient.GetAppByGuid(appGUID)
	stats, _ := cfClient.GetAppStats(appGUID)
	appState := stats["0"].State
	fmt.Println(appGUID + " is " + appState + " with " + strconv.Itoa(app.Instances) + " instances")
}

func canAppReceiveRequests() bool {
	return appCanReceiveRequests
}

func turnAppOn() {
	err := cfClient.StartApp(appGUID)
	if err != nil {
		fmt.Println("Failed to request app to start: ", err)
	} else {
		fmt.Println("Starting app with GUID: " + appGUID)
		app, _ := cfClient.GetAppByGuid(appGUID)
		state := app.State
		fmt.Println("State: ", state)

		var serverStarting = true
		for serverStarting {
			fmt.Print("Checking app state...")
			stats, _ := cfClient.GetAppStats(appGUID)
			appState := stats["0"].State
			fmt.Print(appState + "\n")

			if appState == "RUNNING" {
				fmt.Print("OK! Server started\n")
				serverStarting = false
				appCanReceiveRequests = true
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func turnAppOff() {
	err := cfClient.StopApp(appGUID)
	if err != nil {
		fmt.Println("failed to stop app: ", err)
	} else {
		fmt.Println("App has stopped")
		app, _ := cfClient.GetAppByGuid(appGUID)
		state := app.State
		fmt.Println("State: ", state)
	}
}
