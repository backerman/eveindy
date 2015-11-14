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

// BlueprintsHandlers returns web handler functions that provide information on
// a toon's bluerpints.
func BlueprintsHandlers(localdb db.LocalDB, sde evego.Database, sess server.Sessionizer) (refresh, get web.HandlerFunc) {
	refresh = func(c web.C, w http.ResponseWriter, r *http.Request) {
		s := sess.GetSession(&c, w, r)
		myUserID := s.User
		charID, _ := strconv.Atoi(c.URLParams["charID"])
		apiKeys, err := localdb.APIKeys(myUserID)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Ouch"}`,
				http.StatusInternalServerError)
			return
		}
		// Find the key for this character.
		var myKey *db.XMLAPIKey
		for _, key := range apiKeys {
			for _, toon := range key.Characters {
				if toon.ID == charID {
					myKey = &key
					break
				}
			}
		}
		if myKey == nil {
			if err != nil {
				http.Error(w, `{"status": "Error", "error": "Invalid character supplied."}`,
					http.StatusUnauthorized)
				return
			}
		}

		err = localdb.GetBlueprints(*myKey, charID)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Ouch"}`,
				http.StatusInternalServerError)
			return
		}
		return
	}

	get = func(c web.C, w http.ResponseWriter, r *http.Request) {
		s := sess.GetSession(&c, w, r)
		myUserID := s.User
		charID, err := strconv.Atoi(c.URLParams["charID"])
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Invalid character ID supplied."}`,
				http.StatusBadRequest)
			return
		}
		blueprints, err := localdb.CharacterBlueprints(myUserID, charID)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to access database."}`,
				http.StatusInternalServerError)
			log.Printf("Error accessing database with user %v, character %v: %v", myUserID, charID, err)
			return
		}
		stations := make(map[string]*evego.Station)
		for i := range blueprints {
			bp := &blueprints[i]
			if _, found := stations[strconv.Itoa(bp.StationID)]; !found {
				stn, err := localdb.StationForID(bp.StationID)
				if err == nil {
					stations[strconv.Itoa(bp.StationID)] = stn
				}
			}
		}
		response := struct {
			Blueprints []evego.BlueprintItem     `json:"blueprints"`
			Stations   map[string]*evego.Station `json:"stations"`
		}{blueprints, stations}
		blueprintsJSON, err := json.Marshal(&response)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Unable to marshal JSON."}`,
				http.StatusInternalServerError)
			log.Printf("Error marshalling JSON blueprints with user %v, character %v: %v", myUserID, charID, err)
			return
		}
		w.Write(blueprintsJSON)
		return
	}

	return
}
