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
	"kentix/eliona"
	"kentix/kentix"
	"net/http"
	"time"

	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func initialization() {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, app.AppName())
	defer conn.Close(ctx)

	// Init the app before the first run.
	app.Init(conn, app.AppName(),
		app.ExecSqlFile("conf/init.sql"),
		conf.InitConfiguration,
		eliona.InitEliona,
	)
}

func collectData() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		log.Fatal("conf", "No configs in DB")
		return
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
			log.Info("conf", "Collecting initialized with Configuration %d:\n"+
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

		// Runs the ReadNode. If the current node is currently running, skip the execution
		// After the execution sleeps the configured timeout. During this timeout no further
		// process for this config is started to read the data.
		common.RunOnceWithParam(func(config apiserver.Configuration) {
			log.Info("main", "Collecting %d started", *config.Id)

			collectDataForConfig(config)

			log.Info("main", "Collecting %d finished", *config.Id)

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, *config.Id)
	}
}

func collectDataForConfig(config apiserver.Configuration) {
	deviceInfo, err := kentix.GetDeviceInfo(config)
	if err != nil {
		log.Error("kentix", "getting device info: %v", err)
		return
	}

	if err := eliona.CreateAssetsIfNecessary(config, *deviceInfo); err != nil {
		log.Error("eliona", "creating assets: %v", err)
		return
	}

	if err := eliona.UpsertDeviceInfo(config, *deviceInfo); err != nil {
		log.Error("eliona", "inserting device info: %v", err)
		return
	}

	switch deviceInfo.AssetType {
	case kentix.AlarmManagerAssetType:
	case kentix.AccessPointAssetType:
		doorlocks, err := kentix.GetAccessPointReadings(config)
		if err != nil {
			log.Error("kentix", "getting AccessPoint readings: %v", err)
			return
		}
		for _, doorlock := range doorlocks {
			if err := eliona.CreateDoorlockAssetsIfNecessary(config, doorlock, deviceInfo.Serial); err != nil {
				log.Error("eliona", "creating doorlock assets: %v", err)
				return
			}
			if err := eliona.UpsertDoorlockData(config, doorlock); err != nil {
				log.Error("eliona", "inserting doorlock data: %v", err)
				return
			}
		}
	case kentix.MultiSensorAssetType:
		sensor, err := kentix.GetMultiSensorReadings(config)
		if err != nil {
			log.Error("kentix", "getting MultiSensor readings: %v", err)
			return
		}
		if err := eliona.UpsertMultiSensorData(config, *sensor); err != nil {
			log.Error("eliona", "inserting MultiSensor data: %v", err)
			return
		}
	}
}

// listenApiRequests starts an API server and listen for API requests.
// The API endpoints are defined in the openapi.yaml file.
func listenApiRequests() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"), utilshttp.NewCORSEnabledHandler(
		apiserver.NewRouter(
			apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
			apiserver.NewVersionApiController(apiservices.NewVersionApiService()),
			apiserver.NewCustomizationApiController(apiservices.NewCustomizationApiService()),
		)))
	log.Fatal("main", "Error in API Server: %v", err)
}
