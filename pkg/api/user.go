/*
Copyright © 2014 Brad Ackerman.

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
	"net/http"
	"strconv"

	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji/web"
)

// XMLAPIKeysHandlers returns web handler functions that provide information on
// the user's API keys that have been registered with this application.
func XMLAPIKeysHandlers(localdb db.LocalDB) (list, delete web.HandlerFunc) {
	list = func(c web.C, w http.ResponseWriter, r *http.Request) {
		s := server.GetSession(&c)
		userKeys, err := localdb.APIKeys(s.User)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		userKeysJSON, err := json.Marshal(userKeys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(userKeysJSON)
	}

	delete = func(c web.C, w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This funcition must be called with the POST method",
				http.StatusMethodNotAllowed)
			return
		}
		s := server.GetSession(&c)
		keyID, _ := strconv.Atoi(c.URLParams["keyid"])
		err := localdb.DeleteAPIKey(s.User, keyID)
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
		w.Write([]byte(`{"status": "OK"}`))
	}

	return
}
