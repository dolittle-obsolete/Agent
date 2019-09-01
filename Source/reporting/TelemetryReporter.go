/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Dolittle. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
package reporting

import (
	"agent/log"
	"agent/provisioning"
	"encoding/json"
	"fmt"
)

// TelemetryReporter Represents a system that can report telemetry back to the cloud
type TelemetryReporter struct {
	status    NodeStatus
	providers []ICanProvideTelemetryForNode
}

// NewTelemetryReporter creates a new instance of the TelemetryReporter
func NewTelemetryReporter(provisioner *provisioning.Provider, providers []ICanProvideTelemetryForNode) *TelemetryReporter {
	reporter := new(TelemetryReporter)
	reporter.providers = providers
	reporter.startConfigurationListener(provisioner)
	return reporter
}

func (reporter *TelemetryReporter) startConfigurationListener(provisioner *provisioning.Provider) {
	listener := make(chan provisioning.Node)
	go func() {
		for {
			node := <-listener
			reporter.status.Node = node
		}
	}()
	provisioner.Listen(listener)
}

// ReportCurrentStatus Report the current status of the current node in the current location
func (reporter *TelemetryReporter) ReportCurrentStatus() {
	if !reporter.status.IsValid() {
		log.Informationln("Node configuration is not valid - not reporting.")
		return
	}

	defer func() {
		if reason := recover(); reason != nil {
			log.Errorln("Recovering from panic during ReportCurrentStatus:", reason)
		}
	}()

	log.Informationln("Gathering current status of node")
	reporter.status.Metrics = make(map[string]float32)
	reporter.status.Infos = make(map[string]string)

	for _, provider := range reporter.providers {
		metrics, infos := provider.Provide()

		for _, metric := range metrics {
			reporter.status.Metrics[metric.Type] = metric.Value
		}
		for _, info := range infos {
			reporter.status.Infos[info.Type] = info.Value
		}
	}

	log.Informationln("Sending telemetry to cloud endpoint")

	status, err := json.Marshal(reporter.status)
	if err != nil {
		log.Errorln("Error serializing node status:", err)
		return
	}

	fmt.Println(string(status))
	/*

		body := bytes.NewBuffer(json)
		response, _ := http.Post("https://edge.dolittle.studio/api/Telemetry", "application/json", body)
		//response, _ := http.Post("http://localhost:5000/api/Telemetry", "application/json", body)
		result, _ := ioutil.ReadAll(response.Body)

		fmt.Println(string(result))
	*/
}
