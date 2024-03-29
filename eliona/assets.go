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
	"context"
	"fmt"
	"kentix/apiserver"
	"kentix/conf"
	"kentix/kentix"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func CreateAssetsIfNecessary(config apiserver.Configuration, spec kentix.DeviceInfo) error {
	for _, projectId := range conf.ProjIds(config) {
		if err := createDeviceAssetIfNecessary(config, projectId, spec); err != nil {
			return fmt.Errorf("creating assets for device %s: %v", spec.Serial, err)
		}
	}
	return nil
}

func createDeviceAssetIfNecessary(config apiserver.Configuration, projectId string, spec kentix.DeviceInfo) error {
	assetData := assetData{
		config:        config,
		projectId:     projectId,
		parentAssetId: nil,
		identifier:    spec.Serial,
		assetType:     spec.AssetType,
		name:          fmt.Sprintf("%s (%s)", spec.Name, spec.IPAddress),
		description:   fmt.Sprintf("%s (%s)", spec.Name, spec.Serial),
	}
	return createAssetIfNecessary(assetData)
}

func CreateDoorlockAssetsIfNecessary(config apiserver.Configuration, spec kentix.DoorLock, parentDeviceSerial string) error {
	for _, projectId := range conf.ProjIds(config) {
		parentAssetID, err := conf.GetAssetId(context.Background(), config, projectId, parentDeviceSerial)
		if err != nil {
			return fmt.Errorf("getting parent asset ID: %v", err)
		}
		if err := createDoorlockAssetIfNecessary(config, projectId, parentAssetID, spec); err != nil {
			return fmt.Errorf("creating assets for device %s: %v", spec.Serial, err)
		}
	}
	return nil
}

func createDoorlockAssetIfNecessary(config apiserver.Configuration, projectId string, parentAssetId *int32, spec kentix.DoorLock) error {
	assetData := assetData{
		config:        config,
		projectId:     projectId,
		parentAssetId: parentAssetId,
		identifier:    spec.Serial,
		assetType:     kentix.DoorlockAssetType,
		name:          fmt.Sprintf("%s (%s)", spec.Name, spec.Address),
		description:   fmt.Sprintf("%s (%s)", spec.Name, spec.Serial),
	}
	return createAssetIfNecessary(assetData)
}

type assetData struct {
	config        apiserver.Configuration
	projectId     string
	parentAssetId *int32
	identifier    string
	assetType     string
	name          string
	description   string
}

func createAssetIfNecessary(d assetData) error {
	// Get known asset id from configuration
	assetID, err := conf.GetAssetId(context.Background(), d.config, d.projectId, d.identifier)
	if err != nil {
		return fmt.Errorf("finding asset ID: %v", err)
	}
	if assetID != nil {
		return nil
	}

	newId, err := asset.UpsertAsset(api.Asset{
		ProjectId:               d.projectId,
		GlobalAssetIdentifier:   d.identifier,
		Name:                    *api.NewNullableString(common.Ptr(d.name)),
		AssetType:               d.assetType,
		Description:             *api.NewNullableString(common.Ptr(d.description)),
		ParentFunctionalAssetId: *api.NewNullableInt32(d.parentAssetId),
	})
	if err != nil {
		return fmt.Errorf("upserting asset into Eliona: %v", err)
	}
	if newId == nil {
		return fmt.Errorf("cannot create asset %s", d.name)
	}

	// Remember the asset id for further usage
	if err := conf.InsertSensor(context.Background(), d.config, d.projectId, d.identifier, *newId); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}

	log.Debug("eliona", "Created new asset for project %s and device %s.", d.projectId, d.identifier)

	return nil
}
