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
	"strconv"

	"github.com/backerman/evego"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji/web"
)

// UnusedSalvage returns a web handler function that identifies a user's salvage
// drops that are unused by any blueprint they own.
func UnusedSalvage(localdb db.LocalDB, sde evego.Database, sess server.Sessionizer) web.HandlerFunc {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		s := sess.GetSession(&c, w, r)
		myUserID := s.User
		charID, err := strconv.Atoi(c.URLParams["charID"])
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Invalid character ID supplied."}`,
				http.StatusBadRequest)
			return
		}
		salvage, err := localdb.UnusedSalvage(myUserID, charID)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to access database."}`,
				http.StatusInternalServerError)
			log.Printf("Error accessing database with user %v, character %v: %v", myUserID, charID, err)
			return
		}
		stations := make(map[string]*evego.Station)
		itemInfo := make(map[string]*evego.Item)
		for i := range salvage {
			item := &salvage[i]
			if _, found := stations[strconv.Itoa(item.StationID)]; !found {
				stn, err := localdb.StationForID(item.StationID)
				if err == nil {
					stations[strconv.Itoa(item.StationID)] = stn
				} else {
					// This should really not friggin' happen.
					http.Error(w, `{"status": "Error", "error": "Unable to look up station/outpost."}`,
						http.StatusInternalServerError)
					log.Printf("Unable to look up station/outpost ID %v: %v", item.StationID, err)
					return
				}
			}
			typeIDStr := strconv.Itoa(item.TypeID)
			if _, found := itemInfo[typeIDStr]; !found {
				thisItem, err := sde.ItemForID(item.TypeID)
				if err == nil {
					itemInfo[typeIDStr] = thisItem
				}
			}
		}
		response := struct {
			Items    []evego.InventoryItem     `json:"items"`
			Stations map[string]*evego.Station `json:"stations"`
			ItemInfo map[string]*evego.Item    `json:"itemInfo"`
		}{salvage, stations, itemInfo}
		salvageJSON, err := json.Marshal(&response)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to marshal JSON."}`,
				http.StatusInternalServerError)
			log.Printf("Error marshalling JSON unused salvage with user %v, character %v: %v", myUserID, charID, err)
			return
		}
		w.Write(salvageJSON)
		return

	}
}
