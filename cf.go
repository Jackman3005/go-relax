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
	CFAppGUID    string // cf app my-app-name --guid
	CurrentState AppState
}

var _cfClient *cfclient.Client
func cfClient() *cfclient.Client {
	if _cfClient == nil {
		config := loadConfiguration()
		_cfClient, _= cfclient.NewClient(&config.CFClientConfig)
	}
	return _cfClient
}

func (app *CloudFoundryApp) printSummary() {
	// TODO: Print summary of space, app name, routes, etc.
	cfApp, _ := cfClient().GetAppByGuid(app.CFAppGUID)
	app.updateState()
	fmt.Println(app.CFAppGUID + " is " + string(app.CurrentState) + " with " + strconv.Itoa(cfApp.Instances) + " instances")
}

func (app *CloudFoundryApp) canAppReceiveRequests() bool {
	return app.CurrentState == APP_RUNNING
}

func (app *CloudFoundryApp) updateState() {
	fmt.Print("Checking app state...")
	stats, err := cfClient().GetAppStats(app.CFAppGUID)
	if err != nil {
		if strings.Contains(err.Error(), "CF-AppStoppedStatsError") {
			app.CurrentState = APP_STOPPED
		} else {
			fmt.Println("ERROR: Checking app stats - ", err)
		}
	} else {
		// assuming ["0"] is the zero-eth instance.
		// TODO: Iterate over all instances to see if any are RUNNING
		appState := AppState(stats["0"].State)
		app.CurrentState = appState
	}

	fmt.Print(string(app.CurrentState) + "\n")
}

func (app *CloudFoundryApp) start() {
	fmt.Println("Starting app with GUID: " + app.CFAppGUID)

	err := cfClient().StartApp(app.CFAppGUID)
	if err != nil {
		fmt.Println("Failed to request app to start: ", err)
	} else {
		var serverStarting = true
		for serverStarting {
			app.updateState()

			if app.CurrentState == APP_RUNNING {
				fmt.Print("OK! App is ready to receive requests.\n")
				serverStarting = false
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func (app *CloudFoundryApp) turnAppOff() {
	err := cfClient().StopApp(app.CFAppGUID)
	if err != nil {
		fmt.Println("Failed to request app to stop: ", err)
	} else {
		fmt.Println("App stopped")
	}
}
