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
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/backerman/evego/pkg/character"
	"github.com/backerman/eveindy/pkg/db"
	"github.com/backerman/eveindy/pkg/server"
	"github.com/zenazn/goji/web"
)

// Skill IDs for use in our standings checker.
const (
	connectionsSkillID = 3359
	diplomacySkillID   = 3357
)

// StandingsHandler returns a web handler function that provides information on
// the user's toons' effective standings.
func StandingsHandler(localdb db.LocalDB, sess server.Sessionizer) web.HandlerFunc {
	standingsFunc := func(c web.C, w http.ResponseWriter, r *http.Request) {
		s := sess.GetSession(&c, w, r)
		userID := s.User
		charID, _ := strconv.Atoi(c.URLParams["charID"])
		npcCorpID, _ := strconv.Atoi(c.URLParams["npcCorpID"])
		corpStanding, facStanding, err := localdb.CharacterStandings(userID, charID, npcCorpID)
		if err != nil {
			errorStr := "Unable to get character standings."
			if err == sql.ErrNoRows {
				errorStr = "Invalid corporation ID passed."
			}
			http.Error(w, fmt.Sprintf(`{"status": "Error", "error": "%v"}`, errorStr),
				http.StatusInternalServerError)
			return
		}
		connections, err := localdb.CharacterSkill(userID, charID, connectionsSkillID)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Ouch"}`,
				http.StatusInternalServerError)
			return
		}
		diplomacy, err := localdb.CharacterSkill(userID, charID, diplomacySkillID)
		if err != nil {
			http.Error(w, `{"status": "Error", "error": "Ouch"}`,
				http.StatusInternalServerError)
			return
		}
		effectiveStanding := character.EffectiveStanding(corpStanding, facStanding, connections, diplomacy)
		statusMsg := struct {
			Status            string  `json:"status"`
			EffectiveStanding float64 `json:"standing"`
		}{
			"OK",
			effectiveStanding,
		}

		statusJSON, _ := json.Marshal(&statusMsg)
		w.Write(statusJSON)
		return
	}

	return standingsFunc
}
