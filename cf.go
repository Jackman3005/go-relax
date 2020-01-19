package main

import (
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient"
	"log"
	"strconv"
	"strings"
	"time"
)

type AppState string

const (
	APP_STOPPED       AppState = "STOPPED"
	APP_STARTING      AppState = "STARTING"
	APP_RUNNING       AppState = "RUNNING"
	APP_STATE_UNKNOWN AppState = "UNKNOWN"
)

type CloudFoundryApp struct {
	CFAppGUID            string // cf app my-app-name --guid
	CurrentState         AppState
	LastRequestTimestamp time.Time
}

func NewCloudFoundryApp(cfAppGUID string) CloudFoundryApp {
	return CloudFoundryApp{
		CFAppGUID:            cfAppGUID,
		CurrentState:         APP_STATE_UNKNOWN,
		LastRequestTimestamp: time.Now(),
	}
}

var _cfClient *cfclient.Client

func cfClient() *cfclient.Client {
	if _cfClient == nil {
		config := loadConfiguration()
		_cfClient, _ = cfclient.NewClient(&config.CFClientConfig)
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
	stats, err := cfClient().GetAppStats(app.CFAppGUID)
	if err != nil {
		if strings.Contains(err.Error(), "CF-AppStoppedStatsError") {
			app.CurrentState = APP_STOPPED
		} else {
			log.Fatal("ERROR: Checking app stats - ", err)
		}
	} else {
		// assuming ["0"] is the zero-eth instance.
		// TODO: Iterate over all instances to see if any are RUNNING
		app.CurrentState = AppState(stats["0"].State)
	}
}

func (app *CloudFoundryApp) start() {
	fmt.Println("\nStarting app with GUID: " + app.CFAppGUID)

	err := cfClient().StartApp(app.CFAppGUID)
	if err != nil {
		fmt.Println("Failed to request app to start: ", err)
	} else {
		var serverStarting = true
		fmt.Print("Waiting for app to start.")
		start := time.Now()
		for serverStarting {
			app.updateState()

			if app.CurrentState == APP_RUNNING {
				fmt.Print("OK! App is ready to receive requests.\n")
				end := time.Now()
				fmt.Println("INFO: App startup took ", end.Sub(start))
				serverStarting = false
			} else {
				fmt.Print(".")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func (app *CloudFoundryApp) stop() {
	err := cfClient().StopApp(app.CFAppGUID)
	if err != nil {
		fmt.Println("Failed to request app to stop: ", err)
	} else {
		fmt.Println("App stopped")
	}
}
