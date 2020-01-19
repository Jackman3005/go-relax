package main

import (
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient"
	"strconv"
	"strings"
	"time"
)

type AppState string

const (
	APP_STOPPED  AppState = "STOPPED"
	APP_STARTING AppState = "STARTING"
	APP_RUNNING  AppState = "RUNNING"
)

type CloudFoundryApp struct {
	cfAppGUID    string // cf app my-app-name --guid
	currentState AppState
}

var cfClient, _ = cfclient.NewClient(&cfclient.Config{
	ApiAddress: "https://api.run.pivotal.io",
	Username:   "jackman3000@gmail.com",
	Password:   "1hGGrmZ3bl%uEvgb*2@bNy&92jO6BH",
})

func (app *CloudFoundryApp) printSummary() {
	// TODO: Print summary of space, app name, routes, etc.
	cfApp, _ := cfClient.GetAppByGuid(app.cfAppGUID)
	app.updateState()
	fmt.Println(app.cfAppGUID + " is " + string(app.currentState) + " with " + strconv.Itoa(cfApp.Instances) + " instances")
}

func (app *CloudFoundryApp) canAppReceiveRequests() bool {
	return app.currentState == APP_RUNNING
}

func (app *CloudFoundryApp) updateState() {
	fmt.Print("Checking app state...")
	stats, err := cfClient.GetAppStats(app.cfAppGUID)
	if err != nil {
		if strings.Contains(err.Error(), "CF-AppStoppedStatsError") {
			app.currentState = APP_STOPPED
		} else {
			fmt.Println("ERROR: Checking app stats - ", err)
		}
	} else {
		// assuming ["0"] is the zero-eth instance.
		// TODO: Iterate over all instances to see if any are RUNNING
		appState := AppState(stats["0"].State)
		app.currentState = appState
	}

	fmt.Print(string(app.currentState) + "\n")
}

func (app *CloudFoundryApp) start() {
	fmt.Println("Starting app with GUID: " + app.cfAppGUID)

	err := cfClient.StartApp(app.cfAppGUID)
	if err != nil {
		fmt.Println("Failed to request app to start: ", err)
	} else {
		var serverStarting = true
		for serverStarting {
			app.updateState()

			if app.currentState == APP_RUNNING {
				fmt.Print("OK! App is ready to receive requests.\n")
				serverStarting = false
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func (app *CloudFoundryApp) turnAppOff() {
	err := cfClient.StopApp(app.cfAppGUID)
	if err != nil {
		fmt.Println("Failed to request app to stop: ", err)
	} else {
		fmt.Println("App stopped")
	}
}
