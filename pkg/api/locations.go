/*
Copyright © 2014–5 Brad Ackerman.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/backerman/evego"
	"github.com/zenazn/goji/web"
)

func autocompleteSystems(db evego.Database, search string) *[]evego.SolarSystem {
	// Use make to ensure that we actually have a slice rather than just a nil
	// pointer.
	results := make([]evego.SolarSystem, 0, 5)

	log.Printf("searching %v\n", search)
	systems, _ := db.SolarSystemsForPattern(search + "%")
	for i, s := range systems {
		if i >= 10 {
			break
		}
		results = append(results, s)
	}
	return &results
}

type station struct {
	Name                   string  `json:"name"`
	ID                     int     `json:"id"`
	SystemName             string  `json:"systemName"`
	Security               float64 `json:"security"`
	Constellation          string  `json:"constellation"`
	Region                 string  `json:"region"`
	Outpost                bool    `json:"isOutpost"`
	Owner                  string  `json:"owner"`
	ReprocessingEfficiency float64 `json:"reprocessingEfficiency"`
}

// AutocompleteSystems returns a handler function that serves system
// autocomplete requests.
func AutocompleteSystems(db evego.Database) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		query := c.URLParams["name"]
		systems := &[]evego.SolarSystem{}
		if len(query) >= 3 {
			systems = autocompleteSystems(db, c.URLParams["name"])
		}
		systemsJSON, _ := json.Marshal(*systems)
		w.Write(systemsJSON)
	}
}

// Convert the station/outpost object provided by evego's API into a more
// useful JSON object to be sent to the client.
func stationFromAPI(db evego.Database, s *evego.Station, isOutpost bool) station {
	system, _ := db.SolarSystemForID(s.SystemID)
	stn := station{
		Name:          s.Name,
		ID:            s.ID,
		SystemName:    system.Name,
		Security:      system.Security,
		Constellation: system.Constellation,
		Region:        system.Region,
		Owner:         s.Corporation,
		ReprocessingEfficiency: s.ReprocessingEfficiency,
	}
	if isOutpost {
		stn.Outpost = true
		// Reprocessing efficiency for outposts isn't provided in the SDE,
		// so we default to a basic station.
		stn.ReprocessingEfficiency = 0.50
	}
	return stn
}

func autocompleteStations(db evego.Database, xmlAPI evego.XMLAPI, search string) *[]station {
	// Use make to ensure that we actually have a slice rather than just a nil
	// pointer.
	results := make([]station, 0, 5)
	search = strings.Replace(search, " ", "%", -1)
	stations, _ := db.StationsForName("%" + search + "%")
	outposts, _ := xmlAPI.OutpostsForName("%" + search + "%")

	// Move stations from the API results to our output
	// in alphabetical order. Do not return more than
	// 10 results.
	for len(stations) > 0 || len(outposts) > 0 {
		var s evego.Station
		var isOutpost bool
		numOutposts := len(outposts)
		numStations := len(stations)
		if numStations == 0 ||
			(numOutposts > 0 && (stations)[0].Name > (outposts)[0].Name) {
			s = (outposts)[0]
			outposts = (outposts)[1:]
			isOutpost = true
		} else {
			s = (stations)[0]
			stations = (stations)[1:]
			isOutpost = false
		}

		stn := stationFromAPI(db, &s, isOutpost)
		results = append(results, stn)

		// Stop here if maximum results count has been reached.
		if len(results) == 10 {
			break
		}
	}

	return &results
}

// AutocompleteStations returns a handler function that serves station
// autocomplete requests.
func AutocompleteStations(db evego.Database, xmlAPI evego.XMLAPI) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		query := c.URLParams["name"]
		stations := &[]station{}
		if len(query) >= 3 {
			stations = autocompleteStations(db, xmlAPI, c.URLParams["name"])
		}
		stationsJSON, _ := json.Marshal(*stations)
		w.Write(stationsJSON)
	}
}
