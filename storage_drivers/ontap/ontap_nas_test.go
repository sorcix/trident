// Copyright 2019 NetApp, Inc. All Rights Reserved.

package ontap

import (
	"testing"

	"github.com/stretchr/testify/assert"

	tridentconfig "github.com/netapp/trident/config"
	drivers "github.com/netapp/trident/storage_drivers"
	"github.com/netapp/trident/storage_drivers/ontap/api"
)

func newTestOntapNASDriver(showSensitive *bool) *NASStorageDriver {
	config := &drivers.OntapStorageDriverConfig{}
	sp := func(s string) *string { return &s }

	config.CommonStorageDriverConfig = &drivers.CommonStorageDriverConfig{}
	config.CommonStorageDriverConfig.DebugTraceFlags = make(map[string]bool)
	config.CommonStorageDriverConfig.DebugTraceFlags["method"] = true
	if showSensitive != nil {
		config.CommonStorageDriverConfig.DebugTraceFlags["sensitive"] = *showSensitive
	}

	config.ManagementLIF = "127.0.0.1"
	config.SVM = "SVM1"
	config.Aggregate = "aggr1"
	config.Username = "ontap-nas-user"
	config.Password = "password1!"
	config.StorageDriverName = "ontap-nas"
	config.StoragePrefix = sp("test_")

	nasDriver := &NASStorageDriver{}
	nasDriver.Config = *config

	// ClientConfig holds the configuration data for Client objects
	clientConfig := api.ClientConfig{
		ManagementLIF:           config.ManagementLIF,
		SVM:                     "SVM1",
		Username:                "client_username",
		Password:                "client_password",
		DriverContext:           tridentconfig.DriverContext("driverContext"),
		ContextBasedZapiRecords: 100,
		DebugTraceFlags:         nil,
	}

	nasDriver.API = api.NewClient(clientConfig)
	nasDriver.Telemetry = &Telemetry{
		Plugin:        nasDriver.Name(),
		SVM:           nasDriver.GetConfig().SVM,
		StoragePrefix: *nasDriver.GetConfig().StoragePrefix,
		Driver:        nasDriver,
		done:          make(chan struct{}),
	}

	return nasDriver
}

func TestOntapNasStorageDriverConfigString(t *testing.T) {

	var ontapNasDrivers = []NASStorageDriver{
		*newTestOntapNASDriver(&[]bool{true}[0]),
		*newTestOntapNASDriver(&[]bool{false}[0]),
		*newTestOntapNASDriver(nil),
	}

	sensitiveIncludeList := map[string]string{
		"username":        "ontap-nas-user",
		"password":        "password1!",
		"client username": "client_username",
		"client password": "client_password",
	}

	sensitiveExcludeList := map[string]string{
		"some information": "<REDACTED>",
	}

	externalIncludeList := map[string]string{
		"<REDACTED>":                   "<REDACTED>",
		"username":                     "Username:<REDACTED>",
		"password":                     "Password:<REDACTED>",
		"api":                          "API:<REDACTED>",
		"chap username":                "ChapUsername:<REDACTED>",
		"chap initiator secret":        "ChapInitiatorSecret:<REDACTED>",
		"chap target username":         "ChapTargetUsername:<REDACTED>",
		"chap target initiator secret": "ChapTargetInitiatorSecret:<REDACTED>",
	}

	for _, ontapNasDriver := range ontapNasDrivers {
		sensitive, ok := ontapNasDriver.Config.DebugTraceFlags["sensitive"]

		switch {

		case !ok || (ok && !sensitive):
			for key, val := range externalIncludeList {
				assert.Contains(t, ontapNasDriver.String(), val,
					"ontap-nas driver does not contain %v", key)
				assert.Contains(t, ontapNasDriver.GoString(), val,
					"ontap-nas driver does not contain %v", key)
			}

			for key, val := range sensitiveIncludeList {
				assert.NotContains(t, ontapNasDriver.String(), val,
					"ontap-nas driver contains %v", key)
				assert.NotContains(t, ontapNasDriver.GoString(), val,
					"ontap-nas driver contains %v", key)
			}

		case ok && sensitive:
			for key, val := range sensitiveIncludeList {
				assert.Contains(t, ontapNasDriver.String(), val,
					"ontap-nas driver does not contain %v", key)
				assert.Contains(t, ontapNasDriver.GoString(), val,
					"ontap-nas driver does not contain %v", key)
			}

			for key, val := range sensitiveExcludeList {
				assert.NotContains(t, ontapNasDriver.String(), val,
					"ontap-nas driver redacts %v", key)
				assert.NotContains(t, ontapNasDriver.GoString(), val,
					"ontap-nas driver redacts %v", key)
			}
		}
	}
}
