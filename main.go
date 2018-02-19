//  Copyright (c) 2018 Marty Schoch
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//              http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/chmorgan/nest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var bindAddr = flag.String("addr", ":9264", "http listen address")
var poll = flag.Duration("poll", 2*time.Minute, "poll interval")
var token = flag.String("token", "", "auth token")
var client *nest.Client

var (
	// api
	errCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "nest_api_errors_total",
		Help: "nest_api_errors_total Number of errors encountered",
	})
	// structure
	away = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_structure_away",
		Help: "nest_structure_away Away status (1=away 0=home 2=unkown)",
	}, []string{"structure"})
	// thermostat
	ambientTempC = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_thermostat_ambient_temperature_celcius",
		Help: "nest_thermostat_ambient_temperature_celcius Ambient temperature at the Nest thermostat",
	}, []string{"structure", "device"})
	ambientTempF = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_thermostat_ambient_temperature_fahrenheit",
		Help: "nest_thermostat_ambient_temperature_fahrenheit Ambient temperature at the Nest thermostat",
	}, []string{"structure", "device"})
	humidity = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_thermostat_humidity",
		Help: "nest_thermostat_humidity Humidity at the Nest thermostat",
	}, []string{"structure", "device"})
	targetTempC = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_thermostat_target_temperature_celcius",
		Help: "nest_thermostat_target_temperature_celcius Target temperature at the Nest thermostat",
	}, []string{"structure", "device"})
	targetTempF = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_thermostat_target_temperature_fahrenheit",
		Help: "nest_thermostat_target_temperature_fahrenheit Target temperature at the Nest thermostat",
	}, []string{"structure", "device"})
	hvacState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_thermostat_hvac_state",
		Help: "nest_thermostat_hvac_state Whether HVAC system is actively 1=heating, -1=cooling or is 0=off",
	}, []string{"structure", "device"})
	emergencyHeatState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nest_thermostat_emergency_heat_state",
		Help: "nest_thermostat_emergency_heat_state Emergency Heat status 1=on 0=off",
	}, []string{"structure", "device"})
)

func main() {
	flag.Parse()

	if *token == "" {
		log.Fatal("must supply an auth token")
	}
	client = nest.New("", "STATE", "", "")
	client.Token = *token

	reg := prometheus.NewRegistry()
	reg.Register(away)
	reg.Register(errCount)
	reg.Register(ambientTempC)
	reg.Register(targetTempC)
	reg.Register(ambientTempF)
	reg.Register(targetTempF)
	reg.Register(humidity)
	reg.Register(hvacState)
	reg.Register(emergencyHeatState)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(*bindAddr, nil)

	pollNest()
	ticker := time.NewTicker(*poll)
	for {
		select {
		case <-ticker.C:
			pollNest()
		}
	}
}

func pollNest() {
	structures, err := client.Structures()
	if err != nil {
		errCount.Add(1.0)
		return
	}
	structureIDToName := map[string]string{}
	for k, v := range structures {
		structureIDToName[k] = v.Name
		switch v.Away {
		case "away":
			away.With(prometheus.Labels{"structure": v.Name}).Set(1.0)
		case "home":
			away.With(prometheus.Labels{"structure": v.Name}).Set(0.0)
		default:
			away.With(prometheus.Labels{"structure": v.Name}).Set(2.0)
		}

	}

	devices, err := client.Devices()
	if err != nil {
		errCount.Add(1.0)
		return
	}
	for _, v := range devices.Thermostats {
		ambientTempC.With(prometheus.Labels{
			"structure": structureIDToName[v.StructureID],
			"device":    v.Name,
		}).Set(float64(v.AmbientTemperatureC))
		ambientTempF.With(prometheus.Labels{
			"structure": structureIDToName[v.StructureID],
			"device":    v.Name,
		}).Set(float64(v.AmbientTemperatureF))
		targetTempC.With(prometheus.Labels{
			"structure": structureIDToName[v.StructureID],
			"device":    v.Name,
		}).Set(float64(v.TargetTemperatureC))
		targetTempF.With(prometheus.Labels{
			"structure": structureIDToName[v.StructureID],
			"device":    v.Name,
		}).Set(float64(v.TargetTemperatureF))
		humidity.With(prometheus.Labels{
			"structure": structureIDToName[v.StructureID],
			"device":    v.Name,
		}).Set(float64(v.Humidity))
		switch v.HvacState {
		case "heating":
			hvacState.With(prometheus.Labels{
				"structure": structureIDToName[v.StructureID],
				"device":    v.Name,
			}).Set(1.0)
		case "cooling":
			hvacState.With(prometheus.Labels{
				"structure": structureIDToName[v.StructureID],
				"device":    v.Name,
			}).Set(-1.0)
		default:
			hvacState.With(prometheus.Labels{
				"structure": structureIDToName[v.StructureID],
				"device":    v.Name,
			}).Set(0)
			if v.IsUsingEmergencyHeat {
				emergencyHeatState.With(prometheus.Labels{
					"structure": structureIDToName[v.StructureID],
					"device":    v.Name,
				}).Set(1.0)
			} else {
				emergencyHeatState.With(prometheus.Labels{
					"structure": structureIDToName[v.StructureID],
					"device":    v.Name,
				}).Set(0)
			}
		}
	}
}
