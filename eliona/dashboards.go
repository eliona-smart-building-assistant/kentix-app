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

package eliona

import (
	"fmt"
	"kentix/kentix"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

func KentixDevicesDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Kentix devices"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	multiSensors, _, err := client.NewClient().AssetsAPI.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(kentix.MultiSensorAssetType).
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching MultiSensor assets: %v", err)
	}
	for _, multiSensor := range multiSensors {
		widget := api.Widget{
			WidgetTypeName: "GeneralDisplay",
			AssetId:        multiSensor.Id,
			Details: map[string]interface{}{
				"size":     2,
				"timespan": 7,
			},
			Data: []api.WidgetData{
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "humidity",
						"description":         "Humidity",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "air_quality",
						"description":         "Air quality",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "co2",
						"description":         "CO₂",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         multiSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "temperature",
						"description":         "Temperature",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
			},
		}
		dashboard.Widgets = append(dashboard.Widgets, widget)
	}

	doorlocks, _, err := client.NewClient().AssetsAPI.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(kentix.DoorlockAssetType).
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching Doorlock assets: %v", err)
	}
	for _, doorlock := range doorlocks {
		widget := api.Widget{
			WidgetTypeName: "GeneralDisplay",
			AssetId:        doorlock.Id,
			Details: map[string]interface{}{
				"size":     1,
				"timespan": 7,
			},
			Data: []api.WidgetData{
				{
					ElementSequence: nullableInt32(1),
					AssetId:         doorlock.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "door_contact",
						"description":         "Door contact",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
			},
		}
		dashboard.Widgets = append(dashboard.Widgets, widget)
	}

	return dashboard, nil
}

func nullableInt32(val int32) api.NullableInt32 {
	return *api.NewNullableInt32(common.Ptr(val))
}
