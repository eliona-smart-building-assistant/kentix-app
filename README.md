# Eliona app to access Kentix devices
This [Eliona app for Kentix](https://github.com/eliona-smart-building-assistant/kentix-app) connects the [Kentix devices](https://www.kentix.com/) to an [Eliona](https://www.eliona.io/) enviroment.

This app collects data from Kentix devices such as AccessManager, AlarmManager and MultiSensor and passes their data to Eliona. Each device corresponds to an asset in an Eliona project.

For AccessManager the app discovers all connected smart doorlocks and allows access to their status in Eliona.

NOTE: This app is passing numeric data to Eliona as strings. That was done as a way to overcome inconsistent data formats that the Kentix API provides. This causes that aggregation is not working for the Kentix data. Before deployment, we should change this and work around the inconsistent data formats provided.

## Configuration

The app needs environment variables and database tables for configuration. To edit the database tables the app provides an own API access.

### Registration in Eliona ###

To start and initialize an app in an Eliona environment, the app has to registered in Eliona. For this, an entry in the database table `public.eliona_app` is necessary.

The registration could be done using the reset script.

### Environment variables

- `CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db). Otherwise, the app can't be initialized and started (e.g. `postgres://user:pass@localhost:5432/iot`).

- `INIT_CONNECTION_STRING`: configures the [Eliona database](https://github.com/eliona-smart-building-assistant/go-eliona/tree/main/db) for app initialization like creating schema and tables (e.g. `postgres://user:pass@localhost:5432/iot`). Default is content of `CONNECTION_STRING`.

- `API_ENDPOINT`:  configures the endpoint to access the [Eliona API v2](https://github.com/eliona-smart-building-assistant/eliona-api). Otherwise, the app can't be initialized and started. (e.g. `http://api-v2:3000/v2`)

- `API_TOKEN`: defines the secret to authenticate the app and access the API.

- `API_SERVER_PORT`(optional): defines the port the API server listens on. The default value is `3000`.

- `LOG_LEVEL`(optional): defines the minimum level that should be [logged](https://github.com/eliona-smart-building-assistant/go-utils/blob/main/log/README.md). Not defined the default level is `info`.

### Database tables ###

The app requires configuration data that remains in the database. To do this, the app creates its own database schema `kentix` during initialization. To modify and handle the configuration data the app provides an API access. Have a look at the [API specification](https://eliona-smart-building-assistant.github.io/open-api-docs/?https://raw.githubusercontent.com/eliona-smart-building-assistant/kentix-app/develop/openapi.yaml) how the configuration tables should be used.

- `kentix.configuration`: Configurations for individual Kentix devices. Editable by API.

- `kentix.sensor`: Specific devices, one for each project and configuration. One sensor corresponds to one asset in Eliona.

There is 1:N relationship between configuration and sensor (i.e. one Configuration could be in multiple projects and each would have it's own sensor).

**Generation**: to generate access method to database see Generation section below.


## References

### Kentix App API ###

The Kentix app provides its own API to access configuration data and other functions. The full description of the API is defined in the `openapi.yaml` OpenAPI definition file.

- [API Reference](https://eliona-smart-building-assistant.github.io/open-api-docs/?https://raw.githubusercontent.com/eliona-smart-building-assistant/kentix-app/develop/openapi.yaml) shows details of the API

**Generation**: to generate api server stub see Generation section below.


### Eliona assets ###

This app creates Eliona asset types and attribute sets during initialization.

The data is written for each Kentix device, structured into different subtypes of Elinoa assets. The following subtypes are defined:

- `Input`: Current values reported by Kentix sensors (i.e. MultiSensor readings).
- `Info`: Static data which specifies a Kentix device like address and firmware info.

### Continuous asset creation

All assets are automatically created once the app is run. The old Kentix firmware does not support device discovery, therefore the user must set a Configuration for each device. Then the app creates Eliona assets for that device.

The only exception is AccessManager, which provides the list of connected doorlocks. New ones are then added automatically.

## Tools

### Generate API server stub ###

For the API server the [OpenAPI Generator](https://openapi-generator.tech/docs/generators/openapi-yaml) for go-server is used to generate a server stub. The easiest way to generate the server files is to use one of the predefined generation script which use the OpenAPI Generator Docker image.

```
.\generate-api-server.cmd # Windows
./generate-api-server.sh # Linux
```

### Generate Database access ###

For the database access [SQLBoiler](https://github.com/volatiletech/sqlboiler) is used. The easiest way to generate the database files is to use one of the predefined generation script which use the SQLBoiler implementation. Please note that the database connection in the `sqlboiler.toml` file have to be configured.

```
.\generate-db.cmd # Windows
./generate-db.sh # Linux
```

### Mock Kentix devices ###
`kentix-mock` folder contains mock endpoints implementation:
- `access-manager`: `localhost:3031`
- `alarm-manager`: `localhost:3032`
- `multi-sensor`: `localhost:3033`
