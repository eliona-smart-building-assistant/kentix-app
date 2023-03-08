//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"context"
	"kentix/apiserver"
	"kentix/apiservices"
	"kentix/conf"
	"kentix/kentix"
	"net/http"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func collectData() {
	configs, err := conf.GetConfigs(context.Background())
	if len(configs) <= 0 || err != nil {
		log.Fatal("Kentix", "Couldn't read config from configured database: %v", err)
	}

	for _, config := range configs {
		// Skip config if disabled and set inactive
		if !conf.IsConfigEnabled(config) {
			if conf.IsConfigActive(config) {
				conf.SetConfigActiveState(context.Background(), config, false)
			}
			continue
		}

		// Signals that this config is active
		if !conf.IsConfigActive(config) {
			conf.SetConfigActiveState(context.Background(), config, true)
			log.Info("Kentix", "Collecting initialized with Configuration %d:\n"+
				"Address: %s\n"+
				"API Key: %s\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Active: %t\n"+
				"Project IDs: %v\n",
				*config.Id,
				config.Address,
				config.ApiKey,
				*config.Enable,
				config.RefreshInterval,
				*config.RequestTimeout,
				*config.Active,
				*config.ProjectIDs)
		}

		// Otherwise it would get overwritten with each iteration.
		cc := config

		// Runs the ReadNode. If the current node is currently running, skip the execution
		// After the execution sleeps the configured timeout. During this timeout no further
		// process for this config is started to read the data.
		common.RunOnce(func() {
			log.Info("Kentix", "Collecting %d started", *cc.Id)

			collectDataForConfig(cc)

			log.Info("Kentix", "Collecting %d finished", *cc.Id)

			time.Sleep(time.Second * time.Duration(cc.RefreshInterval))
		}, *cc.Id)
	}
}

func collectDataForConfig(config apiserver.Configuration) {
	deviceInfo, err := kentix.GetDeviceInfo(config)
	if err != nil {
		log.Error("Kentix", "getting device info: %v", err)
		return
	}
	log.Printf(log.InfoLevel, "Kentix", "%+v", deviceInfo)

	// TODO: Verify that this is the correct property to determine device type.
	switch deviceInfo.Type {
	case kentix.AlarmManagerDeviceType:
	case kentix.AccessPointDeviceType:
		// TODO: To be implemented with continuous asset creation.
	case kentix.MultiSensorDeviceType:
		r, err := kentix.GetMultiSensorReadings(config)
		if err != nil {
			log.Error("Kentix", "getting MultiSensor readings: %v", err)
		}
		log.Info("Kentix", "%+v", r)
	}
}

// listenApiRequests starts an API server and listen for API requests.
// The API endpoints are defined in the openapi.yaml file.
func listenApiRequests() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"), apiserver.NewRouter(
		apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
	))
	log.Fatal("Kentix", "Error in API Server: %v", err)
}
