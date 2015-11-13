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
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/backerman/evego"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji/web"
)

// unmarshalKey attempts to unmarshal an XML api key as passed by the client.
// It returns the key iff successful, or a non-nil error otherwise.
func unmarshalKey(r *http.Request, w http.ResponseWriter) (*db.XMLAPIKey, error) {
	keyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read key", http.StatusBadRequest)
		w.Write([]byte(`{"status": "Error"}`))
		return nil, err
	}
	key := db.XMLAPIKey{}
	err = json.Unmarshal(keyJSON, &key)
	if err != nil {
		http.Error(w, "Unable to unmarshal key", http.StatusBadRequest)
		w.Write([]byte(`{"status": "Error"}`))
		return nil, err
	}
	return &key, nil
}

// XMLAPIKeysHandlers returns web handler functions that provide information on
// the user's API keys that have been registered with this application.
func XMLAPIKeysHandlers(localdb db.LocalDB, sess server.Sessionizer) (list, delete, add, refresh web.HandlerFunc) {
	// charRefresh refreshes the characters associated with an API key and returns
	// the current list of characters via the passed responseWriter.
	charRefresh := func(s *db.Session, key *db.XMLAPIKey, w http.ResponseWriter) {
		toons, err := localdb.GetAPICharacters(s.User, *key)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Database connection error (add characters)"}`,
				http.StatusInternalServerError)
			return
		}
		for _, toon := range toons {
			// Update skills for this character.
			err = localdb.GetAPISkills(*key, toon.ID)
			if err != nil {
				http.Error(w, `{"status": "Error", "error": "Database connection error (add skills)"}`,
					http.StatusInternalServerError)
				return
			}

			// Update standings.
			err = localdb.GetAPIStandings(*key, toon.ID)
			if err != nil {
				http.Error(w, `{"status": "Error", "error": "Database connection error (add standings)"}`,
					http.StatusInternalServerError)
				return
			}

			// Update assets.
			err = localdb.GetAssets(*key, toon.ID)
			if err != nil {
				http.Error(w, `{"status": "Error", "error": "Database connection error (add assets)"}`,
					http.StatusInternalServerError)
				log.Printf("Got error in GetAssets: %v", err)
				return
			}

			// Update blueprints.
			err = localdb.GetBlueprints(*key, toon.ID)
			if err != nil {
				http.Error(w, `{"status": "Error", "error": "Database connection error (add blueprints)"}`,
					http.StatusInternalServerError)
				log.Printf("Got error in GetBlueprints: %v", err)
				return
			}

		}
		response := struct {
			Status     string            `json:"status"`
			Characters []evego.Character `json:"characters"`
		}{
			Status:     "OK",
			Characters: toons,
		}
		responseJSON, err := json.Marshal(response)
		w.Write(responseJSON)
		return
	}

	list = func(c web.C, w http.ResponseWriter, r *http.Request) {
		s := sess.GetSession(&c, w, r)
		userKeys, err := localdb.APIKeys(s.User)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		userKeysJSON, err := json.Marshal(userKeys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		w.Write(userKeysJSON)
	}

	delete = func(c web.C, w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This function must be called with the POST method",
				http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		s := sess.GetSession(&c, w, r)
		keyID, _ := strconv.Atoi(c.URLParams["keyid"])
		err := localdb.DeleteAPIKey(s.User, keyID)
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		w.Write([]byte(`{"status": "OK"}`))
		return
	}

	add = func(c web.C, w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This function must be called with the POST method",
				http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		s := sess.GetSession(&c, w, r)
		key, err := unmarshalKey(r, w)
		if err != nil {
			return
		}
		// Ensure that this key is added under the session's user's account.
		key.User = s.User

		err = localdb.AddAPIKey(*key)
		if err != nil {
			http.Error(w, "Database connection error (add key)", http.StatusInternalServerError)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}

		charRefresh(s, key, w)
		return
	}

	refresh = func(c web.C, w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This function must be called with the POST method",
				http.StatusMethodNotAllowed)
			w.Write([]byte(`{"status": "Error"}`))
			return
		}
		s := sess.GetSession(&c, w, r)
		key, err := unmarshalKey(r, w)
		if err != nil {
			return
		}

		// Ensure that this key is added under the session's user's account.
		key.User = s.User
		charRefresh(s, key, w)
		return
	}

	return
}
