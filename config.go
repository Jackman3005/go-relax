package main

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient"
	"os"
)

type Configuration struct {
	CFClientConfig      cfclient.Config `json:"cf_client_config"`     // cf client configuration
	CFAppGUID           string          `json:"app_guid"`             // `cf app my-app-name --guid`
	TargetURL           string          `json:"target_url"`           // URL that this will be a proxy in front of
	InactivityThreshold string          `json:"inactivity_threshold"` // Duration format of go. e.g. 1h10m12s
}

func loadConfiguration() Configuration {
	var config Configuration
	configJson := getEnv("CONFIG_JSON")
	if err := json.Unmarshal([]byte(configJson), &config); err != nil {
		fmt.Println("error while loading configuration", err)
	}
	return config
}

func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return ""
}
