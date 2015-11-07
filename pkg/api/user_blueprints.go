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
	"net/http"
	"strconv"

	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji/web"
)

// BlueprintsHandlers returns web handler functions that provide information on
// a toon's bluerpints.
func BlueprintsHandlers(localdb db.LocalDB, sess server.Sessionizer) (refresh, get web.HandlerFunc) {
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
		// s := sess.GetSession(&c, w, r)
		// myUserID := s.User
		// charID, _ := strconv.Atoi(c.URLParams["charID"])

	}

	return
}
