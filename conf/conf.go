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

package conf

import (
	"context"
	"errors"
	"fmt"
	"kentix/apiserver"
	"kentix/appdb"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var ErrBadRequest = errors.New("bad request")

// InsertConfig inserts or updates
func InsertConfig(ctx context.Context, config apiserver.Configuration) (apiserver.Configuration, error) {
	dbConfig := dbConfigFromApiConfig(config)
	err := dbConfig.InsertG(ctx, boil.Infer())
	if err != nil {
		return apiserver.Configuration{}, err
	}
	return config, err
}

func GetConfig(ctx context.Context, configID int64) (*apiserver.Configuration, error) {
	dbConfig, err := appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(configID),
	).OneG(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching config from database")
	}
	if dbConfig == nil {
		return nil, ErrBadRequest
	}
	apiConfig := apiConfigFromDbConfig(dbConfig)
	return &apiConfig, nil
}

func DeleteConfig(ctx context.Context, configID int64) error {
	count, err := appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(configID),
	).DeleteAllG(ctx)
	if err != nil {
		return fmt.Errorf("fetching config from database")
	}
	if count > 1 {
		return fmt.Errorf("shouldn't happen: deleted more (%v) configs by ID", count)
	}
	if count == 0 {
		return ErrBadRequest
	}
	return nil
}

func dbConfigFromApiConfig(apiConfig apiserver.Configuration) (dbConfig appdb.Configuration) {
	dbConfig.ID = null.Int64FromPtr(apiConfig.Id).Int64
	dbConfig.Address = null.StringFrom(apiConfig.Address)
	dbConfig.APIKey = null.StringFrom(apiConfig.ApiKey)
	dbConfig.Enable = null.BoolFromPtr(apiConfig.Enable)
	dbConfig.RefreshInterval = apiConfig.RefreshInterval
	if apiConfig.RequestTimeout != nil {
		dbConfig.RequestTimeout = *apiConfig.RequestTimeout
	}
	dbConfig.Active = null.BoolFromPtr(apiConfig.Active)
	if apiConfig.ProjectIDs != nil {
		dbConfig.ProjectIds = *apiConfig.ProjectIDs
	}
	return dbConfig
}

func apiConfigFromDbConfig(dbConfig *appdb.Configuration) (apiConfig apiserver.Configuration) {
	apiConfig.Id = &dbConfig.ID
	apiConfig.Address = dbConfig.Address.String
	apiConfig.ApiKey = dbConfig.APIKey.String
	apiConfig.Enable = dbConfig.Enable.Ptr()
	apiConfig.RefreshInterval = dbConfig.RefreshInterval
	apiConfig.RequestTimeout = &dbConfig.RequestTimeout
	apiConfig.Active = dbConfig.Active.Ptr()
	apiConfig.ProjectIDs = common.Ptr[[]string](dbConfig.ProjectIds)
	return apiConfig
}

func apiSensorFromDbSensor(ctx context.Context, dbSensor *appdb.Sensor) (apiSensor apiserver.Sensor, err error) {
	apiSensor.AssetID = dbSensor.AssetID.Ptr()
	dbConfiguration, err := dbSensor.Configuration().OneG(ctx)
	if err != nil {
		return apiSensor, fmt.Errorf("fetching configuration for sensor: %v", err)
	}
	apiSensor.Configuration = apiConfigFromDbConfig(dbConfiguration)
	apiSensor.ProjectID = dbSensor.ProjectID
	apiSensor.SerialNumber = dbSensor.SerialNumber
	return apiSensor, nil
}

func GetConfigs(ctx context.Context) ([]apiserver.Configuration, error) {
	dbConfigs, err := appdb.Configurations().AllG(ctx)
	if err != nil {
		return nil, err
	}
	var apiConfigs []apiserver.Configuration
	for _, dbConfig := range dbConfigs {
		dbConfig.R.GetSensors()
		apiConfigs = append(apiConfigs, apiConfigFromDbConfig(dbConfig))
	}
	return apiConfigs, nil
}

func GetConfigSensors(ctx context.Context, config apiserver.Configuration) ([]apiserver.Sensor, error) {
	if config.Id == nil {
		return nil, fmt.Errorf("shouldn't happen: config ID is null")
	}
	dbSensors, err := appdb.Sensors(
		appdb.SensorWhere.ConfigurationID.EQ(*config.Id),
	).AllG(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up sensors in DB: %v", err)
	}
	if len(dbSensors) == 0 {
		return nil, fmt.Errorf("no sensor found for config %v", config.Id)
	}
	var apiSensors []apiserver.Sensor
	for _, dbSensor := range dbSensors {
		s, err := apiSensorFromDbSensor(ctx, dbSensor)
		if err != nil {
			return nil, fmt.Errorf("creating API sensor from DB sensor: %v", err)
		}
		apiSensors = append(apiSensors, s)
	}
	return apiSensors, nil
}

func GetAssetId(ctx context.Context, config apiserver.Configuration, projId string, deviceId string) (*int32, error) {
	dbSensors, err := appdb.Sensors(
		appdb.SensorWhere.ConfigurationID.EQ(null.Int64FromPtr(config.Id).Int64),
		appdb.SensorWhere.ProjectID.EQ(projId),
		appdb.SensorWhere.SerialNumber.EQ(deviceId),
	).AllG(ctx)
	if err != nil || len(dbSensors) == 0 {
		return nil, err
	}
	return common.Ptr(dbSensors[0].AssetID.Int32), nil
}

func InsertSensor(ctx context.Context, config apiserver.Configuration, projId string, SerialNumber string, assetId int32) error {
	var dbSensor appdb.Sensor
	dbSensor.ConfigurationID = null.Int64FromPtr(config.Id).Int64
	dbSensor.ProjectID = projId
	dbSensor.SerialNumber = SerialNumber
	dbSensor.AssetID = null.Int32From(assetId)
	return dbSensor.InsertG(ctx, boil.Infer())
}

func SetConfigActiveState(ctx context.Context, config apiserver.Configuration, state bool) (int64, error) {
	return appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(null.Int64FromPtr(config.Id).Int64),
	).UpdateAllG(ctx, appdb.M{
		appdb.ConfigurationColumns.Active: state,
	})
}

func ProjIds(config apiserver.Configuration) []string {
	if config.ProjectIDs == nil {
		return []string{}
	}
	return *config.ProjectIDs
}

func IsConfigActive(config apiserver.Configuration) bool {
	return config.Active == nil || *config.Active
}

func IsConfigEnabled(config apiserver.Configuration) bool {
	return config.Enable == nil || *config.Enable
}

func SetAllConfigsInactive(ctx context.Context) (int64, error) {
	return appdb.Configurations().UpdateAllG(ctx, appdb.M{
		appdb.ConfigurationColumns.Active: false,
	})
}
